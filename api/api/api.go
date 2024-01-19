package api

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	luajson "github.com/koeng101/dnadesign/api/api/json"
	"github.com/koeng101/dnadesign/api/gen"
	"github.com/koeng101/dnadesign/lib/bio"
	"github.com/koeng101/dnadesign/lib/bio/fasta"
	"github.com/koeng101/dnadesign/lib/bio/fastq"
	"github.com/koeng101/dnadesign/lib/bio/slow5"
	"github.com/koeng101/dnadesign/lib/primers/pcr"
	"github.com/koeng101/dnadesign/lib/synthesis/codon"
	"github.com/koeng101/dnadesign/lib/synthesis/fix"
	"github.com/koeng101/dnadesign/lib/synthesis/fragment"
	lua "github.com/yuin/gopher-lua"
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

// App implements the dnadesign app
type App struct {
	Router *http.ServeMux
	Logger *slog.Logger
}

// IndexHandler handles the basic HTML page for interacting with polyAPI.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ScalarHandler serves the Scalar API documentation with jsonSwagger
func scalarHandler(jsonSwagger []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Prepare the data for the template. Pass the jsonSwagger as a string.
		data := map[string]interface{}{
			"JsonSwagger": string(jsonSwagger), // Ensure jsonSwagger is in the correct format
		}

		err := templates.ExecuteTemplate(w, "scalar.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
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
	app.Router.HandleFunc("/scalar/", scalarHandler(jsonSwagger))
	app.Router.HandleFunc("/spec.json", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(jsonSwagger) })

	// Lua handlers.
	app.Router.HandleFunc("/api/execute_lua", appImpl.PostExecuteLua)

	// IO handlers.
	app.Router.HandleFunc("/api/io/fasta/parse", appImpl.PostIoFastaParse)
	app.Router.HandleFunc("/api/io/fasta/write", appImpl.PostIoFastaWrite)
	app.Router.HandleFunc("/api/io/genbank/parse", appImpl.PostIoGenbankParse)
	app.Router.HandleFunc("/api/io/genbank/write", appImpl.PostIoGenbankWrite)
	app.Router.HandleFunc("/api/io/fastq/parse", appImpl.PostIoFastqParse)
	app.Router.HandleFunc("/api/io/fastq/write", appImpl.PostIoFastqWrite)
	app.Router.HandleFunc("/api/io/pileup/parse", appImpl.PostIoPileupParse)
	app.Router.HandleFunc("/api/io/pileup/write", appImpl.PostIoPileupWrite)
	app.Router.HandleFunc("/api/io/slow5/parse", appImpl.PostIoSlow5Parse)
	app.Router.HandleFunc("/api/io/slow5/write", appImpl.PostIoSlow5Write)
	app.Router.HandleFunc("/api/io/slow5/svb_compress", appImpl.PostIoSlow5SvbCompress)
	app.Router.HandleFunc("/api/io/slow5/svb_decompress", appImpl.PostIoSlow5SvbDecompress)

	// CDS design handlers.
	app.Router.HandleFunc("/api/cds/fix", appImpl.PostCdsFix)
	app.Router.HandleFunc("/api/cds/optimize", appImpl.PostCdsOptimize)
	app.Router.HandleFunc("/api/cds/translate", appImpl.PostCdsTranslate)

	// PCR handlers.
	app.Router.HandleFunc("/api/pcr/primers/debruijn_barcodes", appImpl.PostPcrPrimersDebruijnBarcodes)
	app.Router.HandleFunc("/api/pcr/primers/marmur_doty", appImpl.PostPcrPrimersMarmurDoty)
	app.Router.HandleFunc("/api/pcr/primers/santa_lucia", appImpl.PostPcrPrimersSantaLucia)
	app.Router.HandleFunc("/api/pcr/primers/melting_temperature", appImpl.PostPcrPrimersMeltingTemperature)
	app.Router.HandleFunc("/api/pcr/primers/design_primers", appImpl.PostPcrPrimersDesignPrimers)
	app.Router.HandleFunc("/api/pcr/complex_pcr", appImpl.PostPcrComplexPcr)
	app.Router.HandleFunc("/api/pcr/simple_pcr", appImpl.PostPcrSimplePcr)

	// Cloning handlers.
	app.Router.HandleFunc("/api/cloning/ligate", appImpl.PostCloningLigate)
	app.Router.HandleFunc("/api/cloning/restriction_digest", appImpl.PostCloningRestrictionDigest)
	app.Router.HandleFunc("/api/cloning/golden_gate", appImpl.PostCloningGoldenGate)
	app.Router.HandleFunc("/api/cloning/fragment", appImpl.PostCloningFragment)

	// Folding handlers.
	app.Router.HandleFunc("/api/folding/zuker", appImpl.PostFoldingZuker)
	app.Router.HandleFunc("/api/folding/linearfold/contra_fold_v2", appImpl.PostFoldingLinearfoldContraFoldV2)
	app.Router.HandleFunc("/api/folding/linearfold/vienna_rna_fold", appImpl.PostFoldingLinearfoldViennaRnaFold)

	// Seqhash handlers.
	app.Router.HandleFunc("/api/seqhash", appImpl.PostSeqhash)
	app.Router.HandleFunc("/api/seqhash_fragment", appImpl.PostSeqhashFragment)

	// Codon Table handlers.
	app.Router.HandleFunc("/api/codon_tables/new", appImpl.PostCodonTablesNew)
	app.Router.HandleFunc("/api/codon_tables/from_genbank", appImpl.PostCodonTablesFromGenbank)
	app.Router.HandleFunc("/api/codon_tables/compromise_tables", appImpl.PostCodonTablesCompromiseTables)
	app.Router.HandleFunc("/api/codon_tables/add_tables", appImpl.PostCodonTablesAddTables)
	app.Router.HandleFunc("/api/codon_tables/default_organisms", appImpl.GetCodonTablesDefaultOrganisms)
	app.Router.HandleFunc("/api/codon_tables/get_organism_table", appImpl.PostCodonTablesGetOrganismTable)

	// Alignment handlers.
	app.Router.HandleFunc("/api/align/needleman_wunsch", appImpl.PostAlignNeedlemanWunsch)
	app.Router.HandleFunc("/api/align/smith_waterman", appImpl.PostAlignSmithWaterman)
	app.Router.HandleFunc("/api/align/mash", appImpl.PostAlignMash)
	app.Router.HandleFunc("/api/align/mash_many", appImpl.PostAlignMashMany)

	// Utils handlers.
	app.Router.HandleFunc("/api/utils/reverse_complement", appImpl.PostUtilsReverseComplement)
	app.Router.HandleFunc("/api/utils/is_palindromic", appImpl.PostUtilsIsPalindromic)

	// Random handlers.
	app.Router.HandleFunc("/api/random/random_dna", appImpl.PostRandomRandomDna)
	app.Router.HandleFunc("/api/random/random_rna", appImpl.PostRandomRandomRna)
	app.Router.HandleFunc("/api/random/random_protein", appImpl.PostRandomRandomProtein)

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

//go:embed json/json.lua
var luaJSON string

func (app *App) ExecuteLua(data string, attachments []gen.Attachment) (string, string, error) {
	L := lua.NewState()
	defer L.Close()
	if err := L.DoString(luaJSON); err != nil {
		panic(err)
	}

	// Add attachments
	luaAttachments := L.NewTable()
	for _, attachment := range attachments {
		luaAttachments.RawSetString(attachment.Name, lua.LString(attachment.Content))
	}
	L.SetGlobal("attachments", luaAttachments)

	// Add IO functions
	L.SetGlobal("fasta_parse", L.NewFunction(app.LuaIoFastaParse))
	L.SetGlobal("genbank_parse", L.NewFunction(app.LuaIoGenbankParse))

	// Add CDS functions
	L.SetGlobal("fix", L.NewFunction(app.LuaCdsFix))
	L.SetGlobal("optimize", L.NewFunction(app.LuaCdsOptimize))
	L.SetGlobal("translate", L.NewFunction(app.LuaCdsTranslate))

	// Add PCR functions
	L.SetGlobal("complex_pcr", L.NewFunction(app.LuaPcrComplexPcr))
	L.SetGlobal("pcr", L.NewFunction(app.LuaPcrSimplePcr))

	// Add Cloning functions
	L.SetGlobal("fragment", L.NewFunction(app.LuaCloningFragment))

	// Add Folding functions
	// Add Seqhash functions
	// Add CodonTable functions
	// Add Utils functions
	// Add Random functions

	// Execute the Lua script
	if err := L.DoString(data); err != nil {
		return "", "", err
	}

	// Extract log and output
	var logBuffer, outputBuffer string
	log := L.GetGlobal("log")
	if str, ok := log.(lua.LString); ok {
		logBuffer = string(str)
	}
	output := L.GetGlobal("output")
	if str, ok := output.(lua.LString); ok {
		outputBuffer = string(str)
	}

	return logBuffer, outputBuffer, nil
}

// luaResponse wraps the core of the lua data -> API calls -> lua data pipeline
func (app *App) luaResponse(L *lua.LState, url, dataString string) int {
	req := httptest.NewRequest("POST", url, strings.NewReader(dataString))
	resp := httptest.NewRecorder()
	app.Router.ServeHTTP(resp, req)

	if resp.Code != 200 {
		L.RaiseError("HTTP request failed: " + resp.Body.String())
		return 0
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	luaData := luajson.DecodeValue(L, data)
	L.Push(luaData)
	return 1
}

// luaStringArrayToGoSlice converts a Lua table at the given index in the Lua stack to a Go slice of strings.
func luaStringArrayToGoSlice(L *lua.LState, index int) ([]string, error) {
	var goStrings []string
	lv := L.Get(index)
	if lv.Type() != lua.LTTable {
		if lv.Type() == lua.LTNil {
			return []string{}, nil
		}
		return nil, fmt.Errorf("argument at index %d is not a table", index)
	}
	tbl := L.ToTable(index)

	tbl.ForEach(func(key lua.LValue, value lua.LValue) {
		if str, ok := value.(lua.LString); ok {
			goStrings = append(goStrings, string(str))
		}
		//else {
		// Handle non-string values if necessary
		//}
	})

	return goStrings, nil
}

/*
*****************************************************************************

# IO functions

*****************************************************************************
*/

func (app *App) PostIoFastaParse(ctx context.Context, request gen.PostIoFastaParseRequestObject) (gen.PostIoFastaParseResponseObject, error) {
	fastaString := *request.Body
	parser := bio.NewFastaParser(strings.NewReader(fastaString + "\n"))
	fastas, err := parser.Parse()
	if err != nil {
		return gen.PostIoFastaParse500TextResponse(fmt.Sprintf("Got error: %s", err)), nil
	}
	data := make([]gen.FastaRecord, len(fastas))
	for i, fastaRecord := range fastas {
		data[i] = gen.FastaRecord(fastaRecord)
	}
	return gen.PostIoFastaParse200JSONResponse(data), nil
}

// LuaIoFastaParse implements fasta_parse in lua.
func (app *App) LuaIoFastaParse(L *lua.LState) int {
	fastaData := L.ToString(1)
	return app.luaResponse(L, "/api/io/fasta/parse", fastaData)
}

func (app *App) PostIoFastaWrite(ctx context.Context, request gen.PostIoFastaWriteRequestObject) (gen.PostIoFastaWriteResponseObject, error) {
	var w bytes.Buffer
	for _, fastaRecord := range *request.Body {
		fastaStruct := fasta.Record(fastaRecord)
		_, _ = fastaStruct.WriteTo(&w) // other than memory problems, there are no circumstances where bytes.Buffer errors
	}
	return gen.PostIoFastaWrite200TextResponse(w.String()), nil
}
func (app *App) LuaIoFastaWrite(L *lua.LState) int { return 0 }

func (app *App) PostIoGenbankParse(ctx context.Context, request gen.PostIoGenbankParseRequestObject) (gen.PostIoGenbankParseResponseObject, error) {
	genbankString := *request.Body
	parser := bio.NewGenbankParser(strings.NewReader(genbankString + "\n"))
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
		data[i] = ConvertGenbankToGenbankRecord(genbankRecord)
	}
	return gen.PostIoGenbankParse200JSONResponse(data), nil
}

// LuaIoGenbankParse implements genbank_parse in lua.
func (app *App) LuaIoGenbankParse(L *lua.LState) int {
	genbankData := L.ToString(1)
	return app.luaResponse(L, "/api/io/genbank/parse", genbankData)
}

func (app *App) PostIoGenbankWrite(ctx context.Context, request gen.PostIoGenbankWriteRequestObject) (gen.PostIoGenbankWriteResponseObject, error) {
	return nil, nil
}
func (app *App) LuaIoGenbankWrite(L *lua.LState) int { return 0 }

func (app *App) PostIoFastqParse(ctx context.Context, request gen.PostIoFastqParseRequestObject) (gen.PostIoFastqParseResponseObject, error) {
	return nil, nil
}
func (app *App) LuaIoFastqParse(L *lua.LState) int { return 0 }

func (app *App) PostIoFastqWrite(ctx context.Context, request gen.PostIoFastqWriteRequestObject) (gen.PostIoFastqWriteResponseObject, error) {
	var w bytes.Buffer
	for _, fastqRead := range *request.Body {
		fastqStruct := fastq.Read{Identifier: fastqRead.Identifier, Sequence: fastqRead.Sequence, Quality: fastqRead.Quality, Optionals: *fastqRead.Optionals}
		_, _ = fastqStruct.WriteTo(&w) // other than memory problems, there are no circumstances where bytes.Buffer errors
	}
	return gen.PostIoFastqWrite200TextResponse(w.String()), nil
}
func (app *App) LuaIoFastqWrite(L *lua.LState) int { return 0 }

func (app *App) PostIoSlow5Parse(ctx context.Context, request gen.PostIoSlow5ParseRequestObject) (gen.PostIoSlow5ParseResponseObject, error) {
	return nil, nil
}
func (app *App) LuaIoSlow5Parse(L *lua.LState) int { return 0 }

func (app *App) PostIoSlow5Write(ctx context.Context, request gen.PostIoSlow5WriteRequestObject) (gen.PostIoSlow5WriteResponseObject, error) {
	var w bytes.Buffer
	var headerValues []slow5.HeaderValue
	requestHeaderValues := request.Body.Header.HeaderValues
	for _, headerValue := range requestHeaderValues {
		headerValues = append(headerValues, slow5.HeaderValue{ReadGroupID: uint32(headerValue.ReadGroupID), Slow5Version: headerValue.Slow5Version, EndReasonHeaderMap: headerValue.EndReasonHeaderMap})
	}
	header := slow5.Header{HeaderValues: headerValues}
	reads := request.Body.Reads
	_, _ = header.WriteTo(&w)
	for _, read := range *reads {
		slow5Struct := ConvertSlow5ReadToRead(read)
		_, _ = slow5Struct.WriteTo(&w)
	}
	return gen.PostIoSlow5Write200TextResponse(w.String()), nil
}
func (app *App) LuaIoSlow5Write(L *lua.LState) int { return 0 }

func (app *App) PostIoSlow5SvbCompress(ctx context.Context, request gen.PostIoSlow5SvbCompressRequestObject) (gen.PostIoSlow5SvbCompressResponseObject, error) {
	return nil, nil
}
func (app *App) LuaIoSlow5SvbCompress(L *lua.LState) int { return 0 }

func (app *App) PostIoSlow5SvbDecompress(ctx context.Context, request gen.PostIoSlow5SvbDecompressRequestObject) (gen.PostIoSlow5SvbDecompressResponseObject, error) {
	return nil, nil
}
func (app *App) LuaIoSlow5SvbDecompress(L *lua.LState) int { return 0 }

func (app *App) PostIoPileupParse(ctx context.Context, request gen.PostIoPileupParseRequestObject) (gen.PostIoPileupParseResponseObject, error) {
	return nil, nil
}
func (app *App) LuaIoPileupParse(L *lua.LState) int { return 0 }

func (app *App) PostIoPileupWrite(ctx context.Context, request gen.PostIoPileupWriteRequestObject) (gen.PostIoPileupWriteResponseObject, error) {
	var w bytes.Buffer
	for _, pileupLine := range *request.Body {
		pileupStruct := ConvertGenPileupLineToPileupLine(pileupLine)
		_, _ = pileupStruct.WriteTo(&w) // other than memory problems, there are no circumstances where bytes.Buffer errors
	}
	return gen.PostIoPileupWrite200TextResponse(w.String()), nil
}
func (app *App) LuaIoPileupWrite(L *lua.LState) int { return 0 }

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
	functions = append(functions, fix.RemoveSequence(sequencesToRemove, "User requested sequence removal"))
	return functions
}

func (app *App) PostCdsFix(ctx context.Context, request gen.PostCdsFixRequestObject) (gen.PostCdsFixResponseObject, error) {
	var ct codon.Table
	organism := string(request.Body.Organism)
	ct, ok := CodonTables[organism]
	if !ok {
		return gen.PostCdsFix400TextResponse(fmt.Sprintf("Organism not found. Got: %s, Need one of the following: %v", organism, []string{"Escherichia coli"})), nil
	}
	sequence, fixChanges, err := fix.Cds(request.Body.Sequence, ct, fixFunctions(request.Body.RemoveSequences))
	if err != nil {
		return gen.PostCdsFix500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	changes := make([]gen.Change, len(fixChanges))
	for i, change := range fixChanges {
		changes[i] = gen.Change{From: change.From, Position: change.Position, Reason: change.Reason, Step: change.Step, To: change.To}
	}
	return gen.PostCdsFix200JSONResponse{Sequence: sequence, Changes: changes}, nil
}

// LuaCdsFix implements fix in lua.
func (app *App) LuaCdsFix(L *lua.LState) int {
	sequence := L.ToString(1)
	organism := L.ToString(2)
	removeSequences, err := luaStringArrayToGoSlice(L, 3)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}

	type fix struct {
		Sequence        string   `json:"sequence"`
		Organism        string   `json:"organism"`
		RemoveSequences []string `json:"remove_sequences"`
	}
	req := &fix{Sequence: sequence, Organism: organism, RemoveSequences: removeSequences}
	b, err := json.Marshal(req)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	return app.luaResponse(L, "/api/cds/fix", string(b))
}

func (app *App) PostCdsOptimize(ctx context.Context, request gen.PostCdsOptimizeRequestObject) (gen.PostCdsOptimizeResponseObject, error) {
	var ct *codon.TranslationTable
	organism := string(request.Body.Organism)
	ct, ok := CodonTables[organism]
	if !ok {
		return gen.PostCdsOptimize400TextResponse(fmt.Sprintf("Organism not found. Got: %s, Need one of the following: %v", organism, []string{"Escherichia coli"})), nil
	}
	var seed int64
	if request.Body.Seed != nil {
		seed = int64(*request.Body.Seed)
	}
	sequence, err := ct.Optimize(request.Body.Sequence, seed)
	if err != nil {
		return gen.PostCdsOptimize500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	return gen.PostCdsOptimize200JSONResponse(sequence), nil
}

// LuaCdsOptimize implements optimize in lua.
func (app *App) LuaCdsOptimize(L *lua.LState) int {
	sequence := L.ToString(1)
	organism := L.ToString(2)
	seed := L.ToInt(3)

	type optimize struct {
		Sequence string `json:"sequence"`
		Organism string `json:"organism"`
		Seed     int    `json:"seed"`
	}
	req := &optimize{Sequence: sequence, Organism: organism, Seed: seed}
	b, err := json.Marshal(req)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	return app.luaResponse(L, "/api/cds/optimize", string(b))
}

func (app *App) PostCdsTranslate(ctx context.Context, request gen.PostCdsTranslateRequestObject) (gen.PostCdsTranslateResponseObject, error) {
	translationTableInteger := request.Body.TranslationTable
	ct := codon.NewTranslationTable(translationTableInteger)
	if ct == nil {
		return gen.PostCdsTranslate500TextResponse(fmt.Sprintf("Translation table of %d not found.", translationTableInteger)), nil
	}
	sequence, err := ct.Translate(request.Body.Sequence)
	if err != nil {
		return gen.PostCdsTranslate500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	return gen.PostCdsTranslate200JSONResponse(sequence), nil
}

// LuaCdsTranslate implements translate in lua.
func (app *App) LuaCdsTranslate(L *lua.LState) int {
	sequence := L.ToString(1)
	translationTable := L.ToInt(2)

	type translate struct {
		Sequence         string `json:"sequence"`
		TranslationTable int    `json:"translation_table"`
	}
	req := &translate{Sequence: sequence, TranslationTable: translationTable}
	b, err := json.Marshal(req)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	return app.luaResponse(L, "/api/cds/translate", string(b))
}

/*
*****************************************************************************

# PCR functions

*****************************************************************************
*/

func (app *App) PostPcrComplexPcr(ctx context.Context, request gen.PostPcrComplexPcrRequestObject) (gen.PostPcrComplexPcrResponseObject, error) {
	var circular bool
	circularPointer := request.Body.Circular
	if circularPointer != nil {
		circular = *circularPointer
	}
	amplicons, err := pcr.Simulate(request.Body.Templates, float64(request.Body.TargetTm), circular, request.Body.Primers)
	if err != nil {
		return gen.PostPcrComplexPcr500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	return gen.PostPcrComplexPcr200JSONResponse(amplicons), nil
}

// LuaPcrComplexPcr implements complex pcr in lua.
func (app *App) LuaPcrComplexPcr(L *lua.LState) int {
	templates, err := luaStringArrayToGoSlice(L, 1)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	primers, err := luaStringArrayToGoSlice(L, 2)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	targetTm := L.ToNumber(3)
	circular := L.ToBool(4)

	req := &gen.PostPcrComplexPcrJSONBody{Circular: &circular, Primers: primers, TargetTm: float32(targetTm), Templates: templates}
	b, err := json.Marshal(req)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	return app.luaResponse(L, "/api/pcr/complex_pcr", string(b))
}

func (app *App) PostPcrSimplePcr(ctx context.Context, request gen.PostPcrSimplePcrRequestObject) (gen.PostPcrSimplePcrResponseObject, error) {
	var circular bool
	circularPointer := request.Body.Circular
	if circularPointer != nil {
		circular = *circularPointer
	}
	amplicons := pcr.SimulateSimple([]string{request.Body.Template}, float64(request.Body.TargetTm), circular, []string{request.Body.ForwardPrimer, request.Body.ReversePrimer})
	if len(amplicons) == 0 {
		return gen.PostPcrSimplePcr500TextResponse("Got no amplicons"), nil
	}
	return gen.PostPcrSimplePcr200JSONResponse(amplicons[0]), nil
}

// LuaPcrSimplePcr implements pcr in lua
func (app *App) LuaPcrSimplePcr(L *lua.LState) int {
	template := L.ToString(1)
	forwardPrimer := L.ToString(2)
	reversePrimer := L.ToString(3)
	targetTm := L.ToNumber(4)
	circular := L.ToBool(5)

	req := &gen.PostPcrSimplePcrJSONBody{Circular: &circular, Template: template, TargetTm: float32(targetTm), ForwardPrimer: forwardPrimer, ReversePrimer: reversePrimer}
	b, err := json.Marshal(req)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	return app.luaResponse(L, "/api/pcr/simple_pcr", string(b))
}

func (app *App) PostPcrPrimersDebruijnBarcodes(ctx context.Context, request gen.PostPcrPrimersDebruijnBarcodesRequestObject) (gen.PostPcrPrimersDebruijnBarcodesResponseObject, error) {
	return nil, nil
}
func (app *App) LuaPcrPrimersDebruijnBarcodes(L *lua.LState) int { return 0 }

func (app *App) PostPcrPrimersMarmurDoty(ctx context.Context, request gen.PostPcrPrimersMarmurDotyRequestObject) (gen.PostPcrPrimersMarmurDotyResponseObject, error) {
	return nil, nil
}
func (app *App) LuaPcrPrimersMarmurDoty(L *lua.LState) int { return 0 }

func (app *App) PostPcrPrimersSantaLucia(ctx context.Context, request gen.PostPcrPrimersSantaLuciaRequestObject) (gen.PostPcrPrimersSantaLuciaResponseObject, error) {
	return nil, nil
}
func (app *App) LuaPcrPrimersSantaLucia(L *lua.LState) int { return 0 }

func (app *App) PostPcrPrimersMeltingTemperature(ctx context.Context, request gen.PostPcrPrimersMeltingTemperatureRequestObject) (gen.PostPcrPrimersMeltingTemperatureResponseObject, error) {
	return nil, nil
}
func (app *App) LuaPcrPrimersMeltingTemperature(L *lua.LState) int { return 0 }

func (app *App) PostPcrPrimersDesignPrimers(ctx context.Context, request gen.PostPcrPrimersDesignPrimersRequestObject) (gen.PostPcrPrimersDesignPrimersResponseObject, error) {
	return nil, nil
}
func (app *App) LuaPcrPrimersDesignPrimers(L *lua.LState) int { return 0 }

/*
*****************************************************************************

# Cloning functions

*****************************************************************************
*/

func (app *App) PostCloningGoldenGate(ctx context.Context, request gen.PostCloningGoldenGateRequestObject) (gen.PostCloningGoldenGateResponseObject, error) {
	return nil, nil
}
func (app *App) LuaCloningGoldenGate(L *lua.LState) int { return 0 }
func (app *App) PostCloningLigate(ctx context.Context, request gen.PostCloningLigateRequestObject) (gen.PostCloningLigateResponseObject, error) {
	return nil, nil
}
func (app *App) LuaCloningLigate(L *lua.LState) int { return 0 }
func (app *App) PostCloningRestrictionDigest(ctx context.Context, request gen.PostCloningRestrictionDigestRequestObject) (gen.PostCloningRestrictionDigestResponseObject, error) {
	return nil, nil
}
func (app *App) LuaCloningRestrictionDigest(L *lua.LState) int { return 0 }
func (app *App) PostCloningFragment(ctx context.Context, request gen.PostCloningFragmentRequestObject) (gen.PostCloningFragmentResponseObject, error) {
	var excludeOverhangs []string
	overhangs := *request.Body.ExcludeOverhangs
	if overhangs != nil {
		excludeOverhangs = overhangs
	}
	fragments, efficiency, err := fragment.Fragment(request.Body.Sequence, request.Body.MinFragmentSize, request.Body.MaxFragmentSize, excludeOverhangs)
	if err != nil {
		return gen.PostCloningFragment500TextResponse(fmt.Sprintf("Got internal server error: %s", err)), nil
	}
	return gen.PostCloningFragment200JSONResponse{Fragments: fragments, Efficiency: float32(efficiency)}, nil
}

// LuaCloningFragment implements fragment in lua.
func (app *App) LuaCloningFragment(L *lua.LState) int {
	sequence := L.ToString(1)
	minFragmentSize := L.ToInt(2)
	maxFragmentSize := L.ToInt(3)
	excludeOverhangs, err := luaStringArrayToGoSlice(L, 4)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}

	type fragmentStruct struct {
		Sequence         string   `json:"sequence"`
		MinFragmentSize  int      `json:"min_fragment_size"`
		MaxFragmentSize  int      `json:"max_fragment_size"`
		ExcludeOverhangs []string `json:"exclude_overhangs"`
	}
	req := &fragmentStruct{Sequence: sequence, MinFragmentSize: minFragmentSize, MaxFragmentSize: maxFragmentSize, ExcludeOverhangs: excludeOverhangs}
	b, err := json.Marshal(req)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	return app.luaResponse(L, "/api/cloning/fragment", string(b))
}

/*
*****************************************************************************

# Folding functions

*****************************************************************************
*/

func (app *App) PostFoldingZuker(ctx context.Context, request gen.PostFoldingZukerRequestObject) (gen.PostFoldingZukerResponseObject, error) {
	return nil, nil
}
func (app *App) LuaFoldingZuker(L *lua.LState) int { return 0 }

func (app *App) PostFoldingLinearfoldContraFoldV2(ctx context.Context, request gen.PostFoldingLinearfoldContraFoldV2RequestObject) (gen.PostFoldingLinearfoldContraFoldV2ResponseObject, error) {
	return nil, nil
}
func (app *App) LuaFoldingLinearfoldContraFoldV2(L *lua.LState) int { return 0 }

func (app *App) PostFoldingLinearfoldViennaRnaFold(ctx context.Context, request gen.PostFoldingLinearfoldViennaRnaFoldRequestObject) (gen.PostFoldingLinearfoldViennaRnaFoldResponseObject, error) {
	return nil, nil
}
func (app *App) LuaFoldingLinearfoldViennaRnaFold(L *lua.LState) int { return 0 }

/*
*****************************************************************************

# Seqhash functions

*****************************************************************************
*/

func (app *App) PostSeqhash(ctx context.Context, request gen.PostSeqhashRequestObject) (gen.PostSeqhashResponseObject, error) {
	return nil, nil
}
func (app *App) LuaSeqhash(L *lua.LState) int { return 0 }

func (app *App) PostSeqhashFragment(ctx context.Context, request gen.PostSeqhashFragmentRequestObject) (gen.PostSeqhashFragmentResponseObject, error) {
	return nil, nil
}
func (app *App) LuaSeqhashFragment(L *lua.LState) int { return 0 }

/*
*****************************************************************************

# CodonTable functions

*****************************************************************************
*/

func (app *App) PostCodonTablesFromGenbank(ctx context.Context, request gen.PostCodonTablesFromGenbankRequestObject) (gen.PostCodonTablesFromGenbankResponseObject, error) {
	return nil, nil
}
func (app *App) LuaCodonTablesFromGenbank(L *lua.LState) int { return 0 }

func (app *App) PostCodonTablesNew(ctx context.Context, request gen.PostCodonTablesNewRequestObject) (gen.PostCodonTablesNewResponseObject, error) {
	return nil, nil
}
func (app *App) LuaCodonTablesNew(L *lua.LState) int { return 0 }

func (app *App) PostCodonTablesCompromiseTables(ctx context.Context, request gen.PostCodonTablesCompromiseTablesRequestObject) (gen.PostCodonTablesCompromiseTablesResponseObject, error) {
	return nil, nil
}
func (app *App) LuaCodonTablesCompromiseTables(L *lua.LState) int { return 0 }

func (app *App) PostCodonTablesAddTables(ctx context.Context, request gen.PostCodonTablesAddTablesRequestObject) (gen.PostCodonTablesAddTablesResponseObject, error) {
	return nil, nil
}
func (app *App) LuaCodonTablesAddTables(L *lua.LState) int { return 0 }

func (app *App) GetCodonTablesDefaultOrganisms(ctx context.Context, request gen.GetCodonTablesDefaultOrganismsRequestObject) (gen.GetCodonTablesDefaultOrganismsResponseObject, error) {
	return nil, nil
}
func (app *App) LuaCodonTablesDefaultOrganisms(L *lua.LState) int { return 0 }

func (app *App) PostCodonTablesGetOrganismTable(ctx context.Context, request gen.PostCodonTablesGetOrganismTableRequestObject) (gen.PostCodonTablesGetOrganismTableResponseObject, error) {
	return nil, nil
}
func (app *App) LuaCodonTablesGetOrganismTable(L *lua.LState) int { return 0 }

/*
*****************************************************************************

# Alignment functions

*****************************************************************************
*/

func (app *App) PostAlignNeedlemanWunsch(ctx context.Context, request gen.PostAlignNeedlemanWunschRequestObject) (gen.PostAlignNeedlemanWunschResponseObject, error) {
	return nil, nil
}
func (app *App) LuaAlignNeedlemanWunsch(L *lua.LState) int { return 0 }

func (app *App) PostAlignSmithWaterman(ctx context.Context, request gen.PostAlignSmithWatermanRequestObject) (gen.PostAlignSmithWatermanResponseObject, error) {
	return nil, nil
}
func (app *App) LuaAlignSmithWaterman(L *lua.LState) int { return 0 }

func (app *App) PostAlignMash(ctx context.Context, request gen.PostAlignMashRequestObject) (gen.PostAlignMashResponseObject, error) {
	return nil, nil
}
func (app *App) LuaPostAlignMash(L *lua.LState) int { return 0 }

func (app *App) PostAlignMashMany(ctx context.Context, request gen.PostAlignMashManyRequestObject) (gen.PostAlignMashManyResponseObject, error) {
	return nil, nil
}
func (app *App) LuaAlignMashMany(L *lua.LState) int { return 0 }

/*
*****************************************************************************

# Utility functions

*****************************************************************************
*/

func (app *App) PostUtilsReverseComplement(ctx context.Context, request gen.PostUtilsReverseComplementRequestObject) (gen.PostUtilsReverseComplementResponseObject, error) {
	return nil, nil
}
func (app *App) Lua(L *lua.LState) int { return 0 }

func (app *App) PostUtilsIsPalindromic(ctx context.Context, request gen.PostUtilsIsPalindromicRequestObject) (gen.PostUtilsIsPalindromicResponseObject, error) {
	return nil, nil
}
func (app *App) LuaUtilsIsPalindromic(L *lua.LState) int { return 0 }

/*
*****************************************************************************

# Random functions

*****************************************************************************
*/

func (app *App) PostRandomRandomDna(ctx context.Context, request gen.PostRandomRandomDnaRequestObject) (gen.PostRandomRandomDnaResponseObject, error) {
	return nil, nil
}
func (app *App) LuaRandomRandomDna(L *lua.LState) int { return 0 }

func (app *App) PostRandomRandomRna(ctx context.Context, request gen.PostRandomRandomRnaRequestObject) (gen.PostRandomRandomRnaResponseObject, error) {
	return nil, nil
}
func (app *App) LuaRandomRandomRna(L *lua.LState) int { return 0 }

func (app *App) PostRandomRandomProtein(ctx context.Context, request gen.PostRandomRandomProteinRequestObject) (gen.PostRandomRandomProteinResponseObject, error) {
	return nil, nil
}
func (app *App) LuaRandomRandomProtein(L *lua.LState) int { return 0 }

/*
*****************************************************************************

# Template for functions

*****************************************************************************
*/
//func (app *App) Lua(L *lua.LState) int { return 0 }
//func (app *App) x(ctx context.Context, request gen.RequestObject) (gen.ResponseObject, error) {
//    return nil, nil
//}
//func (app *App) Lua(L *lua.LState) int { return 0 }
