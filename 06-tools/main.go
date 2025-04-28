package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// MODEL_RUNNER_BASE_URL=http://localhost:12434 go run main.go
func main() {
	ctx := context.Background()

	// Docker Model Runner base URL
	chatURL := os.Getenv("MODEL_RUNNER_BASE_URL") + "/engines/llama.cpp/v1/"
	model := "ai/llama3.2"

	client := openai.NewClient(
		option.WithBaseURL(chatURL),
		option.WithAPIKey(""),
	)

	sayHelloTool := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "say_hello",
			Description: openai.String("Say hello to the given person name"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]string{
						"type": "string",
					},
				},
				"required": []string{"name"},
			},
		},
	}

	vulcanSalute := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "vulcan_salute",
			Description: openai.String("Give a vulcan salute to the given person name"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]string{
						"type": "string",
					},
				},
				"required": []string{"name"},
			},
		},
	}	

	tools := []openai.ChatCompletionToolParam{
		sayHelloTool,
		vulcanSalute,
	}

	userQuestion := openai.UserMessage(`
		Say hello to Jean-Luc Picard 
		and Say hello to James Kirk 
		and make a Vulcan salute to Spock
	`)

	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			userQuestion,
		},
		ParallelToolCalls: openai.Bool(true),
		Tools:             tools,
		Model:             model,
		Temperature:       openai.Opt(0.0),
	}

	// Make completion request
	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		panic(err)
	}

	toolCalls := completion.Choices[0].Message.ToolCalls

	// Return early if there are no tool calls
	if len(toolCalls) == 0 {
		fmt.Println("😡 No function call")
		fmt.Println()
		return
	}

	//params.Messages = append(params.Messages, completion.Choices[0].Message.ToParam())

	for _, toolCall := range toolCalls {
		//fmt.Println(toolCall.Function.Name, toolCall.Function.Arguments)

		switch toolCall.Function.Name {
		case "say_hello":
			args, _ := JsonStringToMap(toolCall.Function.Arguments)
			fmt.Println(sayHello(args))

		default:
			fmt.Println("Unknown function call:", toolCall.Function.Name)
		}
	}

}

func sayHello(arguments map[string]interface{}) string {

	if name, ok := arguments["name"].(string); ok {
		return "Hello " + name
	} else {
		return ""
	}
}

func JsonStringToMap(jsonString string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
