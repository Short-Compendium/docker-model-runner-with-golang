package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// MODEL_RUNNER_BASE_URL=http://localhost:12434  MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:1.5B-F16 go run main.go
// From a container:
// MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest go run main.go
func main() {
	ctx := context.Background()

	// Docker Model Runner base URL
	chatURL := os.Getenv("MODEL_RUNNER_BASE_URL") + "/engines/llama.cpp/v1/"
	model := os.Getenv("MODEL_RUNNER_LLM_CHAT")

	// Create a new OpenAI client
	dmrClient := openai.NewClient(
		option.WithBaseURL(chatURL),
		option.WithAPIKey(""),
	)

	// Start the MCP server process
	cmd := exec.Command(
		"docker",
		"run",
		"-i",
		"--rm",
		"alpine/socat",
		"STDIO",
		"TCP:host.docker.internal:8811",
	)
	// To run it in a container (with compose for example), the image needs to have docker installed

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("😡 Failed to get stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("😡 Failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("😡 Failed to start server: %v", err)
	}
	defer cmd.Process.Kill()

	clientTransport := stdio.NewStdioServerTransportWithIO(stdout, stdin)

	// Create a new MCP client
	mcpClient := mcp_golang.NewClient(clientTransport)

	if _, err := mcpClient.Initialize(ctx); err != nil {
		log.Fatalf("😡 Failed to initialize client: %v", err)
	}

	// Get the list of the available MCP tools
	mcpTools, err := mcpClient.ListTools(ctx, nil)
	if err != nil {
		log.Fatalf("😡 Failed to list tools: %v", err)
	}

	// Convert the mcp tools to openai tools
	openAITools := ConvertToOpenAITools(mcpTools)

	fmt.Println("🛠️  Available Tools (OpenAI format):")
	for _, tool := range openAITools {
		fmt.Println("🔧 Tool:", tool.Function.Name)
		fmt.Println("  - description:", tool.Function.Description)
		fmt.Println("  - parameters:", tool.Function.Parameters)
	}

	// Create a list of messages for the chat completion request
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a pizza expert."),
		//openai.UserMessage("Search information about hawaiian pizza.(only 3 results)"),
		openai.UserMessage(`Search information about hawaiian pizza.(only 3 results)
			Then search information about bananas pizza.(only 1 result)
		`),
	}

	// Create the chat completion parameters
	params := openai.ChatCompletionNewParams{
		Messages:          messages,
		ParallelToolCalls: openai.Bool(true),
		Tools:             openAITools, // ✋ Pass the tools to the request
		Seed:              openai.Int(0),
		Model:             model,
		Temperature:       openai.Opt(0.0),
	}

	// Make initial chat completion request
	completion, err := dmrClient.Chat.Completions.New(ctx, params)
	if err != nil {
		panic(err)
	}

	// Check if the completion contains any tool calls
	detectedToolCalls := completion.Choices[0].Message.ToolCalls

	if len(detectedToolCalls) == 0 {
		fmt.Println("😡 No function call")
		return
	}

	fmt.Println("\n🎉 Detected calls:")

	for _, toolCall := range detectedToolCalls {
		fmt.Println("📣 calling ", toolCall.Function.Name, toolCall.Function.Arguments)

		// toolCall.Function.Arguments is a JSON String
		// Convert the JSON string to a (map[string]any)
		var args map[string]any
		err = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
		if err != nil {
			log.Println("😡 Failed to unmarshal arguments:", err)
		}
		fmt.Println("📝 Arguments:", args)

		// Call the tool with the arguments
		toolResponse, err := mcpClient.CallTool(ctx, toolCall.Function.Name, args)
		if err != nil {
			log.Println("😡 Failed to call tool:", err)
		}
		if toolResponse != nil && len(toolResponse.Content) > 0 && toolResponse.Content[0].TextContent != nil {
			fmt.Println("🎉📝 Tool response:", toolResponse.Content[0].TextContent.Text)
		}
	}

}

func ConvertToOpenAITools(tools *mcp_golang.ToolsResponse) []openai.ChatCompletionToolParam {
	openAITools := make([]openai.ChatCompletionToolParam, len(tools.Tools))

	for i, tool := range tools.Tools {
		schema := tool.InputSchema.(map[string]any)
		openAITools[i] = openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(*tool.Description),
				Parameters: openai.FunctionParameters{
					"type":       "object",
					"properties": schema["properties"],
					"required":   schema["required"],
				},
			},
		}
	}
	return openAITools
}
