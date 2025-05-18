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

// MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal MODEL_RUNNER_LLM_TOOLS=ai/qwen2.5:latest go run main.go

func main() {
	ctx := context.Background()

	// Docker Model Runner base URL
	chatURL := os.Getenv("MODEL_RUNNER_BASE_URL") + "/engines/llama.cpp/v1/"
	modelTools := os.Getenv("MODEL_RUNNER_LLM_TOOLS")
	modelChat := os.Getenv("MODEL_RUNNER_LLM_CHAT")

	fmt.Println("🤖 LLM: ", modelTools)

	// Create a new OpenAI client
	dmrClient := openai.NewClient(
		option.WithBaseURL(chatURL),
		option.WithAPIKey(""),
	)

	systemInstructions := `You are a pizza expert.`
	userQuestion := `
		Search information about hawaiian pizza.(only 3 results)

		Then fetch the URLs from the search information results.

		Make a structured detailed report with all the results,
		The output format MUST be in markdown.
	`

	// Create a new MCP client
	mcpClient, cmd, err := GetMCPClient(ctx)

	if err != nil {
		log.Fatalf("😡 Failed to create MCP client: %v", err)
	}
	defer cmd.Process.Kill()

	// Get the list of the available MCP tools
	mcpTools, err := mcpClient.ListTools(ctx, nil)
	if err != nil {
		log.Fatalf("😡 Failed to list tools: %v", err)
	}

	fmt.Println("🛠️  Available Tools (MCP format): ", len(mcpTools.Tools))

	fmt.Println("⏳ Filtering tools...")

	filteredTools := []mcp_golang.ToolRetType{}
	for _, tool := range mcpTools.Tools {
		if tool.Name == "brave_web_search" || tool.Name == "fetch" { //|| tool.Name == "fetch"
			filteredTools = append(filteredTools, tool)
		}
	}

	fmt.Println("⏳ Converting tools to OpenAI format...")
	// Convert the mcp tools to openai tools
	openAITools := ConvertToOpenAITools(filteredTools)
	for _, tool := range openAITools {
		fmt.Println("🛠️  Tool: ", tool.Function.Name)
		//fmt.Println("🛠️  Description: ", tool.Function.Description)
	}

	// Create a list of messages for the tools and chat completion requests
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemInstructions),
		openai.SystemMessage("Focus only on the part of the text that is related to tools to call."),
		openai.UserMessage(userQuestion),
	}

	DetectToolThenCallIt := func() bool {
		// Create the chat completion parameters
		params := openai.ChatCompletionNewParams{
			Messages:          messages,
			ParallelToolCalls: openai.Bool(true),
			Tools:             openAITools,
			Seed:              openai.Int(0),
			Model:             modelTools,
			Temperature:       openai.Opt(0.0),
		}

		// Make initial chat completion request to detect the tools
		completion, err := dmrClient.Chat.Completions.New(ctx, params)
		if err != nil {
			log.Println("🛠️😡", err)
			return false
		}

		// Check if the completion contains any tool calls
		detectedToolCalls := completion.Choices[0].Message.ToolCalls

		// Exit if no tool calls are detected
		if len(detectedToolCalls) == 0 {
			fmt.Println("👋 No function call")
			return false
		}

		fmt.Println("\n✋ Detected calls:", len(detectedToolCalls))

		toolMessages := make([]openai.ChatCompletionMessageParamUnion, 0, len(detectedToolCalls))
		for _, toolCall := range detectedToolCalls {
			// Call the tool with the arguments
			var args map[string]any
			_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

			fmt.Println("📣 calling ", toolCall.Function.Name, toolCall.Function.Arguments)

			// Call the tool with the arguments
			toolResponse, err := mcpClient.CallTool(ctx, toolCall.Function.Name, args)
			if err != nil {
				log.Println("❌😡 Failed to call tool:", err)
				continue
			}

			// Create a proper tool response message
			toolMessages = append(
				toolMessages,
				openai.ToolMessage(
					toolResponse.Content[0].TextContent.Text,
					toolCall.ID,
				),
			)

			fmt.Println("📝 Tool response:\n", toolResponse.Content[0].TextContent.Text)
		}

		// Add all tool messages at once
		messages = append(messages, toolMessages...)

		return true
	}

	// Loop until all tools are called
	pass := 1
	for DetectToolThenCallIt() {
		fmt.Println("✅ Pass number", pass, "of tool calls executed.")
		pass++
		// to avoid too long and useless searches
		if pass > 2 {
			break
		}
	}

	fmt.Println("🎉 tools execution completed.")

	// only for ai/qwen3:latest
	messages = append(messages, openai.SystemMessage("/no_think"))
		
	params := openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       modelChat,
		Temperature: openai.Opt(0.9),
	}

	stream := dmrClient.Chat.Completions.NewStreaming(ctx, params)

	for stream.Next() {
		chunk := stream.Current()
		// Stream each chunk as it arrives
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}

}

func GetMCPClient(ctx context.Context) (*mcp_golang.Client, *exec.Cmd, error) {

	/*
		cmd := exec.Command(
			"docker",
			"run",
			"-i",
			"--rm",
			"alpine/socat",
			"STDIO",
			"TCP:host.docker.internal:8811",
		)
	*/

	// To run it in a container (with compose for example), the image needs to have docker installed

	// Start the MCP server process

	cmd := exec.Command(
		"socat",
		"STDIO",
		"TCP:host.docker.internal:8811",
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("😡 failed to get stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("😡 failed to get stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("😡 failed to start server: %v", err)
	}

	clientTransport := stdio.NewStdioServerTransportWithIO(stdout, stdin)

	// Create a new MCP client
	mcpClient := mcp_golang.NewClient(clientTransport)

	if _, err := mcpClient.Initialize(ctx); err != nil {
		return nil, nil, fmt.Errorf("😡 failed to initialize client: %v", err)
	}

	return mcpClient, cmd, nil
}

func ConvertToOpenAITools(tools []mcp_golang.ToolRetType) []openai.ChatCompletionToolParam {
	openAITools := make([]openai.ChatCompletionToolParam, len(tools))

	for i, tool := range tools {
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
