package api

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"

	luajson "github.com/koeng101/dnadesign/api/api/json"
	"github.com/koeng101/dnadesign/api/gen"
	lua "github.com/yuin/gopher-lua"
)

func (app *App) ExecuteLua(data string, attachments []gen.Attachment) (string, string, error) {
	L := lua.NewState()
	defer L.Close()

	// Add attachments
	luaAttachments := L.NewTable()
	for _, attachment := range attachments {
		luaAttachments.RawSetString(attachment.Name, lua.LString(attachment.Content))
	}
	L.SetGlobal("attachments", luaAttachments)

	// Add IO functions
	L.SetGlobal("fasta_parse", L.NewFunction(app.LuaIoFastaParse))

	// Add CDS design functions
	L.SetGlobal("fix", L.NewFunction(app.LuaDesignCdsFix))
	L.SetGlobal("optimize", L.NewFunction(app.LuaDesignCdsOptimize))
	L.SetGlobal("translate", L.NewFunction(app.LuaDesignCdsTranslate))

	// Add simulate functions
	L.SetGlobal("fragment", L.NewFunction(app.LuaSimulateFragment))
	L.SetGlobal("complex_pcr", L.NewFunction(app.LuaSimulateComplexPcr))
	L.SetGlobal("pcr", L.NewFunction(app.LuaSimulatePcr))

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
		} else {
			// Handle non-string values if necessary
		}
	})

	return goStrings, nil
}

// LuaIoFastaParse implements fasta_parse in lua.
func (app *App) LuaIoFastaParse(L *lua.LState) int {
	fastaData := L.ToString(1)
	return app.luaResponse(L, "/api/io/fasta/parse", fastaData)
}

// LuaDesignCdsFix implements fix in lua.
func (app *App) LuaDesignCdsFix(L *lua.LState) int {
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
	return app.luaResponse(L, "/api/design/cds/fix", string(b))
}

// LuaDesignCdsOptimize implements optimize in lua.
func (app *App) LuaDesignCdsOptimize(L *lua.LState) int {
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
	return app.luaResponse(L, "/api/design/cds/optimize", string(b))
}

// LuaDesignCdsTranslate implements translate in lua.
func (app *App) LuaDesignCdsTranslate(L *lua.LState) int {
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
	return app.luaResponse(L, "/api/design/cds/translate", string(b))
}

// LuaSimulateFragment implements fragment in lua.
func (app *App) LuaSimulateFragment(L *lua.LState) int {
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
	return app.luaResponse(L, "/api/simulate/fragment", string(b))

}

// LuaSimulateComplexPcr implements complex pcr in lua.
func (app *App) LuaSimulateComplexPcr(L *lua.LState) int {
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

	req := &gen.PostSimulateComplexPcrJSONBody{Circular: &circular, Primers: primers, TargetTm: float32(targetTm), Templates: templates}
	b, err := json.Marshal(req)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	return app.luaResponse(L, "/api/simulate/complex_pcr", string(b))
}

// LuaSimulatePcr implements pcr in lua
func (app *App) LuaSimulatePcr(L *lua.LState) int {
	template := L.ToString(1)
	forwardPrimer := L.ToString(2)
	reversePrimer := L.ToString(3)
	targetTm := L.ToNumber(4)
	circular := L.ToBool(5)

	req := &gen.PostSimulatePcrJSONBody{Circular: &circular, Template: template, TargetTm: float32(targetTm), ForwardPrimer: forwardPrimer, ReversePrimer: reversePrimer}
	b, err := json.Marshal(req)
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	return app.luaResponse(L, "/api/simulate/pcr", string(b))
}
