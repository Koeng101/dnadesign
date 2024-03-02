package api

import (
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

	luajson "github.com/koeng101/dnadesign/api/api/json"
	"github.com/koeng101/dnadesign/api/gen"
	"github.com/koeng101/dnadesign/lib/bio"
	lua "github.com/yuin/gopher-lua"
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
	app.Router.HandleFunc("/chat", chatHandler)

	// Lua handlers.
	app.Router.HandleFunc("/api/execute_lua", appImpl.PostExecuteLua)

	// IO handlers.
	app.Router.HandleFunc("/api/io/fasta/parse", appImpl.PostIoFastaParse)
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
