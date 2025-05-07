package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func generateData(ctx context.Context, client openai.Client, model string, schemaParam openai.ResponseFormatJSONSchemaJSONSchemaParam, question string) (string, error) {

	userQuestion := openai.UserMessage(question)

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
		return "", err
	}
	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}
	if len(completion.Choices[0].Message.Content) == 0 {
		return "", fmt.Errorf("no content returned")
	}
	return completion.Choices[0].Message.Content, nil
}

func getCountriesList(ctx context.Context, client openai.Client, model string, continent string, numberOfCities int) (map[string][]string, error) {

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

	data, err := generateData(ctx, client, model, schemaParam, "List of "+strconv.Itoa(numberOfCities)+" countries in "+continent)
	fmt.Println("Data:", data)

	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	var countriesList map[string][]string

	err = json.Unmarshal([]byte(data), &countriesList)

	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return countriesList, nil
}

func getCountryInformation(ctx context.Context, client openai.Client, model string, country string) (map[string]any, error) {

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

	data, err := generateData(ctx, client, model, schemaParam, fmt.Sprintf("Tell me about %s", country))

	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	var countryInfo map[string]any
	err = json.Unmarshal([]byte(data), &countryInfo)

	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	return countryInfo, nil
}

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
	countriesList, err := getCountriesList(ctx, client, model, "Europe", 5)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Countries List:")
	for idx, country := range countriesList["countries"] {
		fmt.Println(idx, ".", country)
		countryInfo, err := getCountryInformation(ctx, client, model, country)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("  Country Information:")
		fmt.Println("    Name:", countryInfo["name"])
		fmt.Println("    Capital:", countryInfo["capital"])
		fmt.Println("    Languages:", countryInfo["languages"])
		fmt.Println()

	}



}
