package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// MODEL_RUNNER_BASE_URL=http://localhost:12434  MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:0.5B-F16 go run main.go
func main() {
	ctx := context.Background()

	// Docker Model Runner base URL
	chatURL := os.Getenv("MODEL_RUNNER_BASE_URL") + "/engines/llama.cpp/v1/"
	model := os.Getenv("MODEL_RUNNER_LLM_CHAT")

	client := openai.NewClient(
		option.WithBaseURL(chatURL),
		option.WithAPIKey(""),
	)

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type": "string",
			},
			"capital": map[string]any{
				"type": "string",
			},
			"languages": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"name", "capital", "languages"},
	}

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "country_info",
		Description: openai.String("Notable information about a country in the world"),
		Schema:      schema,
		Strict:      openai.Bool(true),
	}

	userQuestion := openai.UserMessage(`
		Tell me about Canada.
	`)

	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			userQuestion,
		},
		Model:       model,
		Temperature: openai.Opt(0.0),
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				JSONSchema: schemaParam,
			},
		},
	}

	// Make completion request
	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		panic(err)
	}
	fmt.Println("Response:", completion.Choices[0].Message.Content)

}
