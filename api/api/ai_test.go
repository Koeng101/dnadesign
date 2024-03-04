package api_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/koeng101/dnadesign/api/api"
	"github.com/koeng101/dnadesign/api/gen"
	"github.com/sashabaranov/go-openai"
)

func TestAiFastaParse(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return
	}
	baseUrl := os.Getenv("BASE_URL")
	model := os.Getenv("MODEL")
	config := openai.DefaultConfig(apiKey)
	if baseUrl != "" {
		config.BaseURL = baseUrl
	}
	client := openai.NewClientWithConfig(config)
	ctx := context.Background()
	userRequest := `I would like you to parse the following FASTA and return to me the headers in a csv file.
The fasta:
>test\nATGC\n>test2\nGATC
`
	message := openai.ChatCompletionMessage{Role: "user", Content: userRequest}
	examples, err := api.RequiredFunctions(ctx, client, model, []openai.ChatCompletionMessage{message})
	if err != nil {
		t.Errorf("Got err: %s", err)
	}
	if examples["fastaParse"] != true {
		t.Errorf("fastaParse != true.")
		for key, value := range examples {
			fmt.Printf("%s:%t\n", key, value)
		}
	}
}

func TestWriteCodeString(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return
	}
	baseUrl := os.Getenv("BASE_URL")
	model := os.Getenv("MODEL")
	config := openai.DefaultConfig(apiKey)
	if baseUrl != "" {
		config.BaseURL = baseUrl
	}
	client := openai.NewClientWithConfig(config)
	ctx := context.Background()
	userRequest := "Please parse the following FASTA and return me a csv. Add headers identifier and sequence to the top of the csv. Data:\n```>test\nATGC\n>test2\nGATC\n"

	message := openai.ChatCompletionMessage{Role: "user", Content: userRequest}
	code, err := api.WriteCodeString(ctx, client, model, []openai.ChatCompletionMessage{message})
	if err != nil {
		t.Errorf("Got error: %s", err)
	}
	// run the code
	output, err := app.ExecuteLua(code, []gen.Attachment{})
	if err != nil {
		t.Errorf("No error should be found. Got err: %s", err)
	}
	expectedOutput := `identifier,sequence
test,ATGC
test2,GATC`
	if strings.TrimSpace(output) != strings.TrimSpace(expectedOutput) {
		t.Errorf("Unexpected response. Expected: " + expectedOutput + "\nGot: " + output)
		fmt.Println(code)
	}
}
