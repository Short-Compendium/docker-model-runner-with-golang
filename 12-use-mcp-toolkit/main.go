package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
)

func main() {
	ctx := context.Background()
	// Start the server process
	cmd := exec.Command(
		"docker",
		"run",
		"-i",
		"--rm",
		"alpine/socat",
		"STDIO",
		"TCP:host.docker.internal:8811",
	)

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
	mcpClient := mcp_golang.NewClient(clientTransport)

	if _, err := mcpClient.Initialize(ctx); err != nil {
		log.Fatalf("😡 Failed to initialize client: %v", err)
	}

	// List available mcpTools
	mcpTools, err := mcpClient.ListTools(ctx, nil)
	if err != nil {
		log.Fatalf("😡 Failed to list tools: %v", err)
	}

	fmt.Println("🛠️ Available MCP Tools:")
	for _, tool := range mcpTools.Tools {

		fmt.Println("🔧 Tool:", tool.Name)
		fmt.Println("  - description:", *tool.Description)

		schema := tool.InputSchema.(map[string]interface{})
		fmt.Println("  - properties:", schema["properties"])
		fmt.Println("  - required:", schema["required"])

	}

	fmt.Println("--------------------------------------------------------------")

	// Arguments for the search tool
	searchArgs := map[string]interface{}{
		"query": "information about hawaiian pizza",
		"max_results": 3,
	}

	fmt.Println("🔎 Calling search tool...")
	
	searchResponse, err := mcpClient.CallTool(ctx, "search", searchArgs)
	
	if err != nil {
		log.Println("😡 Failed to call search tool:", err)
	} else if searchResponse != nil && len(searchResponse.Content) > 0 && searchResponse.Content[0].TextContent != nil {
		fmt.Println("🙂 Search response:", searchResponse.Content[0].TextContent.Text)
	}

}
