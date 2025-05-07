package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)


// MODEL_RUNNER_BASE_URL=http://localhost:12434  MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:1.5B-F16 go run main.go
// From a container:
// MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:1.5B-F16 go run main.go
func main() {
	ctx := context.Background()

	// Docker Model Runner base URL
	chatURL := os.Getenv("MODEL_RUNNER_BASE_URL") + "/engines/llama.cpp/v1/"
	model := os.Getenv("MODEL_RUNNER_LLM_CHAT")

	client := openai.NewClient(
		option.WithBaseURL(chatURL),
		option.WithAPIKey(""),
	)

	// Get a list of countries in Europe
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"countries": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"countries"},
	}

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "List of countries",
		Description: openai.String("List of countries in the world"),
		Schema:      schema,
		Strict:      openai.Bool(true),
	}

	userQuestion := openai.UserMessage("List of 5 countries in Europe")

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

	data := completion.Choices[0].Message.Content

	var countriesList map[string][]string

	err = json.Unmarshal([]byte(data), &countriesList)

	if err != nil {
		panic(err)
	}
	fmt.Println("Countries List:")
	for idx, country := range countriesList["countries"] {
		fmt.Println(idx, ".", country)
	}

}
