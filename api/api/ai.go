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

	"github.com/gorilla/websocket"
	"github.com/sashabaranov/go-openai"
)

// To test multiple times:
// for i in {1..20}; do echo "Run #$i"; go test; done

// OPENAI_API_KEY =""
// MODEL="mistralai/Mixtral-8x7B-Instruct-v0.1"
// BASE_URL="https://api.deepinfra.com/v1/openai"

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

# Examples

*****************************************************************************
*/

//go:embed examples.lua
var examples string

type Example struct {
	Name     string
	Question string
	Text     string
}

// MagicJSONIncantation contains text that somehow consistently gets AI models to return valid JSON.
var MagicJSONIncantation = "Please respond ONLY with valid json that conforms to this json_schema: %s\n. Do not include additional text other than the object json as we will load this object with json.loads()."

// FunctionExamples is a complete list of DnaDesign lua examples.
var FunctionExamples = parseExamples(examples)

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
func RequiredFunctions(ctx context.Context, client *openai.Client, model string, userRequest string) (map[string]bool, error) {
	var resultMap map[string]bool
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					// You will be answering questions about a user request, but not directly answering the user request
					Content: fmt.Sprintf(MagicJSONIncantation, FunctionExamplesJSONSchema),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(`USER REQUEST: %s`, userRequest),
				},
			},
			Stop: []string{"}"}, // this stop token makes sure nothing else is generated.
		},
	)
	if err != nil {
		return resultMap, err
	}
	if len(resp.Choices) == 0 {
		return resultMap, fmt.Errorf("Got zero responses")
	}
	response := resp.Choices[0].Message.Content + "}" // add back stop token
	err = json.Unmarshal([]byte(response), &resultMap)
	if err != nil {
		fmt.Println(response)
		return resultMap, err
	}

	return resultMap, nil
}
