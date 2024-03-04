package api_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/koeng101/dnadesign/api/api"
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
>test\nATGC\ntest2\nGATC
`
	examples, err := api.RequiredFunctions(ctx, client, model, userRequest)
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
