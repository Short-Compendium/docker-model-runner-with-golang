package main

import (
	"context"
	"embeddings-demo/rag"
	"fmt"
	"log"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var chunks = []string{
	`# The Avengers
	"The Avengers" is a classic British spy-fi television series that aired from 1961 to 1969. 
	The show exemplifies the unique style of 1960s British television with its blend of espionage,
	 science fiction, and quintessentially British humor. 
	The series follows secret agents working for a specialized branch of British intelligence, 
	battling eccentric villains and foiling bizarre plots to undermine national security.`,

	`# John Steed
    John Steed, portrayed by Patrick Macnee, is the quintessential English gentleman spy 
	who never leaves home without his trademark bowler hat and umbrella (which conceals various weapons). 
	Charming, witty, and deceptively dangerous, Steed approaches even the most perilous situations 
	with impeccable manners and a dry sense of humor. 
	His refined demeanor masks his exceptional combat skills and razor-sharp intelligence.`,

	`# Emma Peel
     Emma Peel, played by Diana Rigg, is perhaps the most iconic of Steed's partners. 
	 A brilliant scientist, martial arts expert, and fashion icon, Mrs. Peel combines beauty, brains, 
	 and remarkable fighting skills. Clad in her signature leather catsuits, she represents the modern, 
	 liberated woman of the 1960s. Her name is a play on "M-appeal" (man appeal), 
	 but her character transcended this origin to become a feminist icon.`,

	`# Tara King
     Tara King, played by Linda Thorson, was Steed's final regular partner in the original series. 
	 Younger and somewhat less experienced than her predecessors, King was nevertheless a trained agent 
	 who continued the tradition of strong female characters. 
	 Her relationship with Steed had more romantic undertones than previous partnerships, 
	 and she brought a fresh, youthful energy to the series.`,

	`# Mother
    Mother, portrayed by Patrick Newell, is Steed's wheelchair-bound superior who appears in later seasons. 
	Operating from various unusual locations, this eccentric spymaster directs operations with a mix of authority 
	and peculiarity that fits perfectly within the show's offbeat universe.`,
}

// MODEL_RUNNER_BASE_URL=http://localhost:12434 go run main.go
func main() {
	ctx := context.Background()

	llmURL := os.Getenv("MODEL_RUNNER_BASE_URL") + "/engines/llama.cpp/v1/"
	embeddingsModel := "ai/mxbai-embed-large"
	chatModel := "ai/qwen2.5:0.5B-F16"

	client := openai.NewClient(
		option.WithBaseURL(llmURL),
		option.WithAPIKey(""),
	)

	// -------------------------------------------------
	// Create a vector store
	// -------------------------------------------------
	store := rag.MemoryVectorStore{
		Records: make(map[string]rag.VectorRecord),
	}

	// -------------------------------------------------
	// Create and save the embeddings from the chunks
	// -------------------------------------------------
	fmt.Println("⏳ Creating the embeddings...")

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
			_, errSave := store.Save(rag.VectorRecord{
				Prompt:    chunk,
				Embedding: embeddingsResponse.Data[0].Embedding,
			})
			if errSave != nil {
				fmt.Println("😡:", errSave)
			}
		}
	}

	fmt.Println("✋", "Embeddings created, total of records", len(store.Records))
	fmt.Println()

	// -------------------------------------------------
	// Search for similarities
	// -------------------------------------------------
	//userQuestion := "Tell me about the English series called The Avengers?"
	// userQuestion := "Who is John Steed?"
	userQuestion := "Who is Emma Peel?"
	// userQuestion := "Who is Tara King?"
	// userQuestion := "Who is Mother?"

	fmt.Println("⏳ Searching for similarities...")

	// -------------------------------------------------
	// Create embedding from the user question
	// -------------------------------------------------
	embeddingsResponse, err := client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(userQuestion),
		},
		Model: embeddingsModel,
	})
	if err != nil {
		log.Fatal("😡:", err)
	}

	// -------------------------------------------------
	// Create a vector record from the user embedding
	// -------------------------------------------------
	embeddingFromUserQuestion := rag.VectorRecord{
		Embedding: embeddingsResponse.Data[0].Embedding,
	}

	similarities, _ := store.SearchTopNSimilarities(embeddingFromUserQuestion, 0.6, 2)

	documentsContent := "Documents:\n"

	for _, similarity := range similarities {
		fmt.Println("✅ CosineSimilarity:", similarity.CosineSimilarity, "Chunk:", similarity.Prompt)
		documentsContent += similarity.Prompt
	}
	documentsContent += "\n"
	fmt.Println("✋", "Similarities found, total of records", len(similarities))
	fmt.Println()

	// -------------------------------------------------
	// Generate completion
	// -------------------------------------------------
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(`You are a useful AI agent expert with TV series. 
		Use only the following documents to answer:`),
		openai.SystemMessage(documentsContent),
		openai.UserMessage(userQuestion),
	}

	param := openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       chatModel,
		Temperature: openai.Opt(0.0),
	}

	stream := client.Chat.Completions.NewStreaming(ctx, param)

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

	fmt.Println()
}
