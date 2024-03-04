package api

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sashabaranov/go-openai"
	"github.com/xeipuuv/gojsonschema"
)

// To test multiple times:
// for i in {1..20}; do echo "Run #$i"; go test; done

// API_KEY=""
// MODEL="mistralai/Mixtral-8x7B-Instruct-v0.1"
// BASE_URL="https://api.deepinfra.com/v1/openai"
// CODE_MODEL="Phind/Phind-CodeLlama-34B-v2"

// API_KEY=""
// MODEL="gpt-4-0125-preview"
// BASE_URL="https://api.openai.com/v1"

/*
*****************************************************************************

# Chat functions

*****************************************************************************
*/

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 1024 * 16,
	WriteBufferSize: 1024 * 1024 * 16,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Chat struct {
	Type    string     `json:"type"`
	Content string     `json:"content"`
	Files   []ChatFile `json:"files"`
}

type ChatFile struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	apiKey := os.Getenv("API_KEY")
	baseUrl := os.Getenv("BASE_URL")
	model := os.Getenv("MODEL")
	config := openai.DefaultConfig(apiKey)
	if baseUrl != "" {
		config.BaseURL = baseUrl
	}
	client := openai.NewClientWithConfig(config)
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		var chats []Chat
		err = json.Unmarshal(p, &chats)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %s", err)
			return
		}
		var messages []openai.ChatCompletionMessage
		for _, chat := range chats {
			messages = append(messages, openai.ChatCompletionMessage{Role: chat.Type, Content: chat.Content})
		}

		stream, err := client.CreateChatCompletionStream(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    model,
				Messages: messages,
				Stream:   true,
			},
		)
		if err != nil {
			fmt.Printf("ChatCompletionStream error: %v\n", err)
			return
		}
		defer stream.Close()
		for {
			var response openai.ChatCompletionStreamResponse
			response, err = stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				fmt.Printf("\nStream error: %v\n", err)
				break
			}
			fmt.Printf(response.Choices[0].Delta.Content)

			_ = conn.WriteMessage(messageType, []byte(response.Choices[0].Delta.Content))
		}

	}
}

/*
*****************************************************************************

# Question asker

*****************************************************************************
*/

// MagicJSONIncantation contains text that somehow consistently gets AI models to return valid JSON.
var MagicJSONIncantation = "Please respond ONLY with valid json that conforms to this json_schema: %s\n. Do not include additional text other than the object json as we will load this object with json.loads()."

func AskQuestions(ctx context.Context, client *openai.Client, model string, jsonSchema string, messages []openai.ChatCompletionMessage) (map[string]bool, error) {
	schemaLoader := gojsonschema.NewStringLoader(jsonSchema)

	var resultMap map[string]bool
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(MagicJSONIncantation, jsonSchema),
	})
	createJson := func(messages []openai.ChatCompletionMessage, model string) (openai.ChatCompletionResponse, error) {
		return client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    model,
				Messages: messages,
				Stop:     []string{"}"}, // this stop token makes sure nothing else is generated.
			},
		)
	}
	resp, err := createJson(messages, model)
	if err != nil {
		return resultMap, err
	}
	if len(resp.Choices) == 0 {
		return resultMap, fmt.Errorf("Got zero responses")
	}
	response := resp.Choices[0].Message.Content + "}" // add back stop token
	response = ParseFromLLM(response)
	documentLoader := gojsonschema.NewStringLoader(response)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return resultMap, err
	}
	if result.Valid() {
		err = json.Unmarshal([]byte(response), &resultMap)
		if err != nil {
			fmt.Printf("Failed to parse this response: ```json\n%s\n```\n", response)
			return resultMap, err
		}
		return resultMap, nil
	}
	return resultMap, fmt.Errorf("response not valid to schema")
}

func toolJSONSchema() string {
	type tool struct {
		Name     string
		Question string
	}

	tools := []tool{
		tool{Name: "code", Question: "The user is asking for code to be written"},
	}

	properties := make(map[string]interface{})
	required := []string{}

	for _, ex := range tools {
		// Generating each property's schema
		properties[ex.Name] = map[string]interface{}{
			"type":        "boolean",
			"description": ex.Question,
		}
		// Accumulating required properties
		required = append(required, ex.Name)
	}

	// Putting together the final schema
	schema := map[string]interface{}{
		"$schema":              "http://json-schema.org/draft-07/schema#", // Assuming draft-07; adjust if necessary
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}

	jsonBytes, _ := json.Marshal(schema)
	return string(jsonBytes)
}

/*
*****************************************************************************

# Examples

*****************************************************************************
*/

//go:embed examples.lua
var Examples string

type Example struct {
	Name     string
	Question string
	Text     string
}

// FunctionExamples is a complete list of DnaDesign lua examples.
var FunctionExamples = parseExamples(Examples)

// FunctionExamplesJSONSchema is a JSON schema containing questions of whether
// or not a given user request requires a
var FunctionExamplesJSONSchema = generateJSONSchemaFromExamples(FunctionExamples)

func parseExamples(content string) []Example {
	var examples []Example
	startPattern := regexp.MustCompile(`-- START: (\w+)`)
	questionPattern := regexp.MustCompile(`-- QUESTION: ([^\n]+)`)
	endPattern := regexp.MustCompile(`-- END`)

	startMatches := startPattern.FindAllStringSubmatchIndex(content, -1)
	endIndexes := endPattern.FindAllStringIndex(content, -1)

	for i, matches := range startMatches {
		name := content[matches[2]:matches[3]]
		sectionStart := matches[1]
		sectionEnd := endIndexes[i][0]

		sectionContent := content[sectionStart:sectionEnd]

		questionMatch := questionPattern.FindStringSubmatchIndex(sectionContent)
		question := ""
		var textStartPos int
		if questionMatch != nil {
			// Include the question in the extract, but not in the Text
			question = content[sectionStart+questionMatch[2] : sectionStart+questionMatch[3]]
			textStartPos = sectionStart + questionMatch[1] + len("\n") // Adjust to start text after the question line
		} else {
			textStartPos = sectionStart
		}

		text := content[textStartPos:sectionEnd]
		text = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(text, "") // Trim leading/trailing whitespace

		examples = append(examples, Example{Name: name, Question: question, Text: text})
	}

	return examples
}

func generateJSONSchemaFromExamples(examples []Example) string {
	properties := make(map[string]interface{})
	required := []string{}

	for _, ex := range examples {
		// Generating each property's schema
		properties[ex.Name] = map[string]interface{}{
			"type":        "boolean",
			"description": ex.Question,
		}
		// Accumulating required properties
		required = append(required, ex.Name)
	}

	// Putting together the final schema
	schema := map[string]interface{}{
		"$schema":              "http://json-schema.org/draft-07/schema#", // Assuming draft-07; adjust if necessary
		"type":                 "object",
		"properties":           properties,
		"required":             required,
		"additionalProperties": false,
	}

	jsonBytes, _ := json.Marshal(schema)
	return string(jsonBytes)
}

// RequiredFunctions takes in a userRequest and returns a map of the examples
// that should be inserted along with that request to generate lua code.
func RequiredFunctions(ctx context.Context, client *openai.Client, model string, messages []openai.ChatCompletionMessage) (map[string]bool, error) {
	return AskQuestions(ctx, client, model, generateJSONSchemaFromExamples(parseExamples(Examples)), messages)
}

func RequiredFunctionsText(ctx context.Context, client *openai.Client, model string, messages []openai.ChatCompletionMessage) (string, error) {
	resultMap, err := RequiredFunctions(ctx, client, model, messages)
	if err != nil {
		return "", err
	}

	// Now that we have the required functions, get their content
	exampleText := ``
	examples := parseExamples(Examples)
	for _, example := range examples {
		_, ok := resultMap[example.Name]
		if ok {
			exampleText = exampleText + example.Text + "\n"
		}
	}
	return exampleText, nil
}

func RequiredFunctionsTextWithRetry(ctx context.Context, client *openai.Client, model string, messages []openai.ChatCompletionMessage, maxAttempts int) (string, error) {
	var lastErr error
	for attempts := 0; attempts < maxAttempts; attempts++ {
		exampleText, err := RequiredFunctionsText(ctx, client, model, messages)
		if err == nil {
			return exampleText, err
		}
		lastErr = err
	}
	return "", lastErr
}

/*
*****************************************************************************

# Code writing

*****************************************************************************
*/

// MagicLuaIncantation contains text that gets the AI models to return valid lua.
var MagicLuaIncantation = "Please response ONLY with valid lua. The following functions are loaded into the sandbox: \n```lua\n%s\n```\n. Do not use any values from the examples except the functions. Do not write or read files, only print data."

// WriteCode takes in user messages and writes code to accomplish the specific tasks.
func WriteCode(ctx context.Context, client *openai.Client, model string, messages []openai.ChatCompletionMessage, exampleText string) (*openai.ChatCompletionStream, error) {
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf(MagicLuaIncantation, exampleText),
	})
	return client.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
			Stream:   true,
		},
	)
}

func WriteCodeString(ctx context.Context, client *openai.Client, model string, messages []openai.ChatCompletionMessage, exampleText string) (string, error) {
	stream, err := WriteCode(ctx, client, model, messages, exampleText)
	if err != nil {
		return "", err
	}
	defer stream.Close()
	var buffer strings.Builder
	for {
		var response openai.ChatCompletionStreamResponse
		response, err = stream.Recv()
		if errors.Is(err, io.EOF) {
			err = nil
			break
		}

		if err != nil {
			break
		}
		buffer.WriteString(response.Choices[0].Delta.Content)
	}

	return ParseFromLLM(buffer.String()), err
}

func ParseFromLLM(input string) string {
	luaPrefix := "```lua"
	jsonPrefix := "```json"
	codePrefix := "```"
	codeSuffix := "```"

	// Check for ```lua ... ```
	luaStartIndex := strings.Index(input, luaPrefix)
	if luaStartIndex != -1 {
		luaEndIndex := strings.Index(input[luaStartIndex+len(luaPrefix):], codeSuffix)
		if luaEndIndex != -1 {
			return input[luaStartIndex+len(luaPrefix) : luaStartIndex+len(luaPrefix)+luaEndIndex]
		}
	}

	// Check for ```json ... ```
	jsonStartIndex := strings.Index(input, jsonPrefix)
	if jsonStartIndex != -1 {
		jsonEndIndex := strings.Index(input[jsonStartIndex+len(jsonPrefix):], codeSuffix)
		if jsonEndIndex != -1 {
			return input[jsonStartIndex+len(jsonPrefix) : jsonStartIndex+len(jsonPrefix)+jsonEndIndex]
		}
	}

	// Check for ``` ... ```
	codeStartIndex := strings.Index(input, codePrefix)
	if codeStartIndex != -1 {
		codeEndIndex := strings.Index(input[codeStartIndex+len(codePrefix):], codeSuffix)
		if codeEndIndex != -1 {
			return input[codeStartIndex+len(codePrefix) : codeStartIndex+len(codePrefix)+codeEndIndex]
		}
	}

	// Return original if no markers found
	return input
}
