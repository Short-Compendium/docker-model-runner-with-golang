package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go"
)

// MODEL_RUNNER_BASE_URL=http://localhost:12434 go run main.go
func main() {
	// Docker Model Runner Chat base URL
	llmURL := os.Getenv("MODEL_RUNNER_BASE_URL") + "/engines/llama.cpp/v1/"
	model := os.Getenv("MODEL_RUNNER_LLM_CHAT")
	//model := "ai/qwen2.5:0.5B-F16"
	//model := "ai/qwen2.5:1.5B-F16"
	//model := "ai/qwen2.5:latest"

	client := openai.NewClient(
		option.WithBaseURL(llmURL),
		option.WithAPIKey(""),
	)

	ctx := context.Background()

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a useful AI agent expert with TV series."),
		openai.UserMessage("Tell me about the English series called The Avengers?"),
	}

	param := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    model,
		Temperature: openai.Opt(0.8),
	}

	completion, err := client.Chat.Completions.New(ctx, param)

	if err != nil {
		log.Fatalln("😡:", err)
	}
	fmt.Println(completion.Choices[0].Message.Content)

	// Print a hello world message
	
}
