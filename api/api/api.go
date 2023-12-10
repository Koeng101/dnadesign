package api

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/flowchartsman/swaggerui"
	"github.com/koeng101/dnadesign/api/gen"
	"github.com/koeng101/dnadesign/bio"
	"github.com/koeng101/dnadesign/primers/pcr"
	"github.com/koeng101/dnadesign/synthesis/codon"
	"github.com/koeng101/dnadesign/synthesis/fix"
	"github.com/koeng101/dnadesign/synthesis/fragment"
)

//go:embed codon_tables/freqB.json
var ecoliCodonTable []byte

var (
	// CodonTables is a list of default codon tables.
	CodonTables = map[string]*codon.TranslationTable{
		"Escherichia coli": codon.ParseCodonJSON(ecoliCodonTable),
	}
)

//go:embed templates/*
var templatesFS embed.FS
var templates = template.Must(template.ParseFS(templatesFS, "templates/*"))

// IndexHandler handles the basic HTML page for interacting with polyAPI.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// App implements the dnadesign app
type App struct {
	Router *http.ServeMux
	Logger *slog.Logger
}

func corsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Set CORS header
		handler.ServeHTTP(w, r)
	})
}

// initializeApp starts here
func InitializeApp() App {
	var app App
	app.Router = http.NewServeMux()

	// Initiate logger
	app.Logger = slog.Default()
	appImpl := gen.NewStrictHandler(&app, nil)

	// Handle swagger docs.
	swagger, err := gen.GetSwagger()
	if err != nil {
		log.Fatalf("Failed to get swagger: %s", err)
	}
	jsonSwagger, err := swagger.MarshalJSON()
	if err != nil {
		log.Fatalf("Failed to marshal swagger: %s", err)
	}

	// Handle static files
	subFS, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		log.Fatal(err)
	}

	// Handle routes.
	app.Router.HandleFunc("/", indexHandler)
	app.Router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFS))))
	app.Router.Handle("/swagger/", http.StripPrefix("/swagger", swaggerui.Handler(jsonSwagger)))
	app.Router.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) { w.Write(jsonSwagger) })

	// Lua handlers.
	app.Router.HandleFunc("/api/execute_lua", appImpl.PostExecuteLua)

	// IO handlers.
	app.Router.HandleFunc("/api/io/fasta/parse", appImpl.PostIoFastaParse)
	app.Router.HandleFunc("/api/io/genbank/parse", appImpl.PostIoGenbankParse)

	// CDS design handlers.
	app.Router.HandleFunc("/api/design/cds/fix", appImpl.PostDesignCdsFix)
	app.Router.HandleFunc("/api/design/cds/optimize", appImpl.PostDesignCdsOptimize)
	app.Router.HandleFunc("/api/design/cds/translate", appImpl.PostDesignCdsTranslate)

	// Simulate handlers.
	app.Router.HandleFunc("/api/simulate/fragment", appImpl.PostSimulateFragment)
	app.Router.HandleFunc("/api/simulate/complex_pcr", appImpl.PostSimulateComplexPcr)
	app.Router.HandleFunc("/api/simulate/pcr", appImpl.PostSimulatePcr)

	return app
}

/*
*****************************************************************************

# Lua functions

*****************************************************************************
*/

func (app *App) PostExecuteLua(ctx context.Context, request gen.PostExecuteLuaRequestObject) (gen.PostExecuteLuaResponseObject, error) {
	script := request.Body.Script
	attachments := *request.Body.Attachments
	output, log, err := app.ExecuteLua(script, attachments)
	if err != nil {
		return gen.PostExecuteLua500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	return gen.PostExecuteLua200JSONResponse{Output: output, Log: log}, nil
}

/*
*****************************************************************************

# IO functions

*****************************************************************************
*/

func (app *App) PostIoFastaParse(ctx context.Context, request gen.PostIoFastaParseRequestObject) (gen.PostIoFastaParseResponseObject, error) {
	fastaString := *request.Body
	parser, err := bio.NewFastaParser(strings.NewReader(fastaString + "\n"))
	if err != nil {
		return gen.PostIoFastaParse500TextResponse(fmt.Sprintf("Got error: %s", err)), nil
	}
	fastas, err := parser.Parse()
	if err != nil {
		return gen.PostIoFastaParse500TextResponse(fmt.Sprintf("Got error: %s", err)), nil
	}
	data := make([]gen.FastaRecord, len(fastas))
	for i, fastaRecord := range fastas {
		data[i] = gen.FastaRecord(*fastaRecord)
	}
	return gen.PostIoFastaParse200JSONResponse(data), nil
}

func (app *App) PostIoGenbankParse(ctx context.Context, request gen.PostIoGenbankParseRequestObject) (gen.PostIoGenbankParseResponseObject, error) {
	genbankString := *request.Body
	parser, err := bio.NewGenbankParser(strings.NewReader(genbankString + "\n"))
	if err != nil {
		return gen.PostIoGenbankParse500TextResponse(fmt.Sprintf("Got error: %s", err)), nil
	}
	genbanks, err := parser.Parse()
	if err != nil {
		return gen.PostIoGenbankParse500TextResponse(fmt.Sprintf("Got error: %s", err)), nil
	}
	data := make([]gen.GenbankRecord, len(genbanks))
	for i, genbankRecord := range genbanks {
		err := genbankRecord.StoreFeatureSequences()
		if err != nil {
			return gen.PostIoGenbankParse500TextResponse(fmt.Sprintf("Got error: %s", err)), nil
		}
		data[i] = gen.GenbankRecord(*genbankRecord)
	}
	return gen.PostIoGenbankParse200JSONResponse(data), nil
}

/*
*****************************************************************************

# CDS functions

*****************************************************************************
*/

func fixFunctions(sequencesToRemove []string) []func(string, chan fix.DnaSuggestion, *sync.WaitGroup) {
	var functions []func(string, chan fix.DnaSuggestion, *sync.WaitGroup)
	functions = append(functions, fix.RemoveSequence([]string{"AAAAAAAA", "GGGGGGGG"}, "Homopolymers removed. Interferes with synthesis."))
	functions = append(functions, fix.RemoveRepeat(18))
	functions = append(functions, fix.GcContentFixer(0.80, 0.20))
	functions = append(functions, fix.RemoveSequence([]string{"GGTCTC", "GAAGAC", "CACCTGC"}, "Common TypeIIS restriction enzymes - BsaI, BbsI, PaqCI"))
	return functions
}

func (app *App) PostDesignCdsFix(ctx context.Context, request gen.PostDesignCdsFixRequestObject) (gen.PostDesignCdsFixResponseObject, error) {
	var ct codon.Table
	organism := string(request.Body.Organism)
	ct, ok := CodonTables[organism]
	if !ok {
		return gen.PostDesignCdsFix400TextResponse(fmt.Sprintf("Organism not found. Got: %s, Need one of the following: %v", organism, []string{"Escherichia coli"})), nil
	}
	sequence, fixChanges, err := fix.Cds(request.Body.Sequence, ct, fixFunctions(request.Body.RemoveSequences))
	if err != nil {
		return gen.PostDesignCdsFix500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	changes := make([]gen.Change, len(fixChanges))
	for i, change := range fixChanges {
		changes[i] = gen.Change{From: change.From, Position: change.Position, Reason: change.Reason, Step: change.Step, To: change.To}
	}
	return gen.PostDesignCdsFix200JSONResponse{Sequence: sequence, Changes: changes}, nil
}

func (app *App) PostDesignCdsOptimize(ctx context.Context, request gen.PostDesignCdsOptimizeRequestObject) (gen.PostDesignCdsOptimizeResponseObject, error) {
	var ct *codon.TranslationTable
	organism := string(request.Body.Organism)
	ct, ok := CodonTables[organism]
	if !ok {
		return gen.PostDesignCdsOptimize400TextResponse(fmt.Sprintf("Organism not found. Got: %s, Need one of the following: %v", organism, []string{"Escherichia coli"})), nil
	}
	var seed int
	if request.Body.Seed != nil {
		seed = *request.Body.Seed
	}
	ct.UpdateWeights(ct.AminoAcids)
	sequence, err := ct.Optimize(request.Body.Sequence, seed)
	if err != nil {
		return gen.PostDesignCdsOptimize500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	return gen.PostDesignCdsOptimize200JSONResponse(sequence), nil
}

func (app *App) PostDesignCdsTranslate(ctx context.Context, request gen.PostDesignCdsTranslateRequestObject) (gen.PostDesignCdsTranslateResponseObject, error) {
	translationTableInteger := request.Body.TranslationTable
	ct := codon.NewTranslationTable(translationTableInteger)
	if ct == nil {
		return gen.PostDesignCdsTranslate500TextResponse(fmt.Sprintf("Translation table of %d not found.", translationTableInteger)), nil
	}
	sequence, err := ct.Translate(request.Body.Sequence)
	if err != nil {
		return gen.PostDesignCdsTranslate500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	return gen.PostDesignCdsTranslate200JSONResponse(sequence), nil
}

/*
*****************************************************************************

# Simulation begins here.

*****************************************************************************
*/

func (app *App) PostSimulateFragment(ctx context.Context, request gen.PostSimulateFragmentRequestObject) (gen.PostSimulateFragmentResponseObject, error) {
	var excludeOverhangs []string
	overhangs := *request.Body.ExcludeOverhangs
	if overhangs != nil {
		excludeOverhangs = overhangs
	}
	fragments, efficiency, err := fragment.Fragment(request.Body.Sequence, request.Body.MinFragmentSize, request.Body.MaxFragmentSize, excludeOverhangs)
	if err != nil {
		return gen.PostSimulateFragment500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	return gen.PostSimulateFragment200JSONResponse{Fragments: fragments, Efficiency: float32(efficiency)}, nil
}
func (app *App) PostSimulateComplexPcr(ctx context.Context, request gen.PostSimulateComplexPcrRequestObject) (gen.PostSimulateComplexPcrResponseObject, error) {
	var circular bool
	circularPointer := request.Body.Circular
	if circularPointer != nil {
		circular = *circularPointer
	}
	amplicons, err := pcr.Simulate(request.Body.Templates, float64(request.Body.TargetTm), circular, request.Body.Primers)
	if err != nil {
		return gen.PostSimulateComplexPcr500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	return gen.PostSimulateComplexPcr200JSONResponse(amplicons), nil
}
func (app *App) PostSimulatePcr(ctx context.Context, request gen.PostSimulatePcrRequestObject) (gen.PostSimulatePcrResponseObject, error) {
	var circular bool
	circularPointer := request.Body.Circular
	if circularPointer != nil {
		circular = *circularPointer
	}
	amplicons := pcr.SimulateSimple([]string{request.Body.Template}, float64(request.Body.TargetTm), circular, []string{request.Body.ForwardPrimer, request.Body.ReversePrimer})
	if len(amplicons) == 0 {
		return gen.PostSimulatePcr500TextResponse("Got no amplicons"), nil
	}
	return gen.PostSimulatePcr200JSONResponse(amplicons[0]), nil
}
func (app *App) PostSimulateGoldengate(ctx context.Context, request gen.PostSimulateGoldengateRequestObject) (gen.PostSimulateGoldengateResponseObject, error) {
	return nil, nil
}
func (app *App) PostSimulateLigate(ctx context.Context, request gen.PostSimulateLigateRequestObject) (gen.PostSimulateLigateResponseObject, error) {
	return nil, nil
}
func (app *App) PostSimulateRestrictionDigest(ctx context.Context, request gen.PostSimulateRestrictionDigestRequestObject) (gen.PostSimulateRestrictionDigestResponseObject, error) {
	return nil, nil
}
