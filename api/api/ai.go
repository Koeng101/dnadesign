package api

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/sashabaranov/go-openai"
)

// To test multiple times:
// for i in {1..20}; do echo "Run #$i"; go test; done

// OPENAI_API_KEY =""
// MODEL="mistralai/Mixtral-8x7B-Instruct-v0.1"
// BASE_URL="https://api.deepinfra.com/v1/openai"

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
