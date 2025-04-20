package main

import (
	"context"
	"embeddings-demo/rag"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var chunks = []string{
	`Lions run in the savannah`,
	`Birds fly in the sky`,
	`Frogs swim in the pond`,
	`Fish swim in the sea`,
}

// MODEL_RUNNER_BASE_URL=http://localhost:12434 go run main.go
func main() {
	ctx := context.Background()

	llmURL := os.Getenv("MODEL_RUNNER_BASE_URL") + "/engines/llama.cpp/v1/"
	embeddingsModel := "ai/mxbai-embed-large"

	client := openai.NewClient(
		option.WithBaseURL(llmURL),
		option.WithAPIKey(""),
	)

	// -------------------------------------------------
	// Generate embeddings from user question
	// -------------------------------------------------
	userQuestion := "Which animals swim?"

	fmt.Println("⏳ Creating embeddings from user question...")

	embeddingsFromUserQuestion, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(userQuestion),
		},
		Model: embeddingsModel,
	})
	if err != nil {
		fmt.Println(err)
	}

	// -------------------------------------------------
	// Generate embeddings from chunks
	// -------------------------------------------------
	fmt.Println("⏳ Creating embeddings from chunks...")

	for _, chunk := range chunks {
		embeddingsResponse, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
			Input: openai.EmbeddingNewParamsInputUnion{
				OfString: openai.String(chunk),
			},
			Model: embeddingsModel,
		})

		if err != nil {
			fmt.Println(err)
		} else {
			cosineSimilarity := rag.CosineSimilarity(
				embeddingsResponse.Data[0].Embedding,
				embeddingsFromUserQuestion.Data[0].Embedding,
			)
			fmt.Println("🔗 Cosine similarity with ", chunk, "=", cosineSimilarity)
		}
	}
}
