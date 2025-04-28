package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	helloTool := openai.ChatCompletionToolParam{
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

	tools := []openai.ChatCompletionToolParam{
		helloTool,
	}

	systemInstructions := openai.SystemMessage(`You are a useful AI agent.`)

	systemToolsInstructions := openai.SystemMessage(` 
	Your job is to understand the user prompt and decide if you need to use tools to run external commands.
	Ignore all things not related to the usage of a tool
	`)

	userQuestion := openai.UserMessage(`Say hello to Jean-Luc Picard and Say hello to James Kirk and Spock.
	Then generate a nice output with the results and insert fancy emojis between each hello.
	`)

	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			systemInstructions,
			systemToolsInstructions,
			userQuestion,
		},
		ParallelToolCalls: openai.Bool(true),
		Tools:             tools,
		Seed:              openai.Int(0),
		Model:             model,
		Temperature:       openai.Opt(0.0),
	}

	// Make initial completion request
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

	firstCompletionResult := "RESULTS:\n"

	for _, toolCall := range toolCalls {
		//fmt.Println(toolCall.Function.Name, toolCall.Function.Arguments)

		switch toolCall.Function.Name {
		case "say_hello":
			args, _ := JsonStringToMap(toolCall.Function.Arguments)
			firstCompletionResult += sayHello(args) + "\n"
			//fmt.Println(sayHello(args))

		default:
			fmt.Println("Unknown function call:", toolCall.Function.Name)
		}
	}

	systemToolsInstructions = openai.SystemMessage(` 
	If you detect that the user prompt is related to a tool, 
	ignore this part and focus on the other parts.
	`)

	params = openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			systemInstructions,
			systemToolsInstructions,
			openai.SystemMessage(firstCompletionResult),
			userQuestion,
		},
		//ParallelToolCalls: openai.Bool(true),
		//Tools:             tools,
		//Seed:              openai.Int(0),
		Model:       model,
		Temperature: openai.Opt(0.8),
	}

	stream := client.Chat.Completions.NewStreaming(ctx, params)

	for stream.Next() {
		chunk := stream.Current()
		// Stream each chunk as it arrives
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}

	if err := stream.Err(); err != nil {
		log.Fatalln("😡:", err)
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
