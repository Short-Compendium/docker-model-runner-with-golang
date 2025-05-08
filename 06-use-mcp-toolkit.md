# Boosting Docker Model Runner with Docker MCP Toolkit
> or how to use MCP servers with LLMs simply and securely.

**Disclaimer:** The goal of this blog post is not to explain MCP (for that you can read [Understanding the Model Context Protocol (MCP)](https://k33g.hashnode.dev/understanding-the-model-context-protocol-mcp)).

Today I want to show how we can use an MCP server (with STDIO transport) in a generative AI application, all in Go.

- For the generative AI part, I'll use **Docker Model Runner** (You can read an introduction to DMR here: [First Contact with Docker Model Runner in Golang](https://k33g.hashnode.dev/first-contact-with-docker-model-runner-in-golang?source=more_series_bottom_blogs)).
- For the MCP part, I'll use **Docker MCP Toolkit**, which I'll talk about in the first part of this blog post.

## Docker MCP Toolkit

Before talking about **Docker MCP Toolkit**, I should first introduce the **Docker MCP Catalog**. The **Docker MCP Catalog** is integrated with Docker Hub and serves as a starting point to discover a set of popular containerized MCP servers. The goal is to facilitate the development of generative AI applications. **Docker MCP Catalog** provides centralized access to official and trusted MCP tools (e.g., Elastic, Neo4j, Heroku, ...).

With introductions made, we can now move to **Docker MCP Toolkit**.

**Docker MCP Toolkit** is an extension for **Docker Desktop** that:

- Simplifies the installation and management of MCP servers
- **Manages credentials (secrets, tokens, ...) securely** (with secure storage of credentials)
- Applies access control
- Secures the runtime environment
- Offers one-click connection with popular MCP clients like Gordon (Docker AI Agent), Claude, Cursor, VSCode, ...

> Coming soon, users will be able to create and share their own MCP servers on Docker Hub.

But what interests me today is that you can use **Docker MCP Toolkit** with your own MCP clients (your generative AI application, for example).

The application I want to develop today should be able to understand from a prompt that I'm looking for information about **Hawaiian pizzas** (and why not other types of pizzas) and search for this information on the web.

So everything will start from a phrase like this: `"Search information about hawaiian pizza."` 🍍🥓


## First contact: installing an MCP server

First, I need an MCP server that knows how to search the internet and return results.

### Install the Docker MCP Toolkit extension

First, you'll need to install the **Docker MCP Toolkit** extension in Docker Desktop:

![Docker MCP Toolkit](/blog-imgs/mcp-toolkit-00.png)


### Find and install the DuckDuckGo MCP server

Once the extension is installed, you can search through the list of available MCP servers. For our example, we'll use the **DuckDuckGo** MCP server, which offers search functionality:

![MCP DuckDuckGo](/blog-imgs/mcp-toolkit-01.png)

If you click on it, you can see the details of the tools offered:

![MCP DuckDuckGo](/blog-imgs/mcp-toolkit-02.png)

And by activating the switch to the right of the MCP server name, you'll install this server in Docker Desktop:

![MCP DuckDuckGo](/blog-imgs/mcp-toolkit-03.png)

From now on, you'll find it in the list of installed MCP servers:

![MCP DuckDuckGo](/blog-imgs/mcp-toolkit-08.png)

## First use

Now, thanks to **Docker MCP Toolkit**, your MCP clients can access all the "tools" of the installed MCP servers by specifying this command:

```bash
docker run -i --rm alpine/socat STDIO TCP:host.docker.internal:8811
```

**Docker MCP Toolkit** acts somewhat like a proxy to the MCP servers, and it's as if the MCP client only sees one MCP server.

OK, but today I want to use this MCP server from my Go code. For that, I'll use the **[mcp-golang](https://mcpgolang.com/introduction)** project, which is a framework for developing MCP clients and servers.

### List of tools offered by the DuckDuckGo MCP server

To get the list of available tools, I'll need to:
1. Start the MCP server
2. Initialize the STDIO transport
3. Create an MCP client
4. Initialize the MCP client (connect it to the server)
5. Request the list of available tools

My Go code will look like this:

```golang
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
```

And when I run it, I'll get this:

```raw
🛠️ Available MCP Tools:
🔧 Tool: search
  - description: 
    Search DuckDuckGo and return formatted results.

    Args:
        query: The search query string
        max_results: Maximum number of results to return (default: 10)
        ctx: MCP context for logging
    
  - properties: map[max_results:map[default:10 title:Max Results type:integer] query:map[title:Query type:string]]
  - required: [query]
🔧 Tool: fetch_content
  - description: 
    Fetch and parse content from a webpage URL.

    Args:
        url: The webpage URL to fetch content from
        ctx: MCP context for logging
    
  - properties: map[url:map[title:Url type:string]]
  - required: [url]
```

So I have two tools at my disposal, `search` and `fetch_content`. The one I'm interested in is `search`. Let's see how to call it.

> Note: if I had installed other MCP servers, I would have a longer list of tools.


### Executing the search tool of the DuckDuckGo MCP server

To execute a tool, the Go code is extremely simple:

1. I build the arguments to send to the server (my search text and the number of expected results)
2. I then call the `search` tool, passing it these arguments
3. I display the results

```golang
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
```

And when I run it, I'll get this:
```raw
🔎 Calling search tool...
🙂 Search response: Found 3 search results:

1. Hawaiian pizza - Wikipedia
   URL: https://en.wikipedia.org/wiki/Hawaiian_pizza
   Summary: Hawaiianpizzais apizzainvented in Canada, topped with pineapple, tomato sauce, mozzarella cheese, and either ham or bacon. History. Sam Panopoulos, a Greek-born Canadian, created the firstHawaiianpizzaat the Satellite Restaurant in Chatham-Kent, Ontario, Canada, in 1962.

2. Here Are The Facts About Hawaiian Pizza - Mashed
   URL: https://www.mashed.com/230119/here-are-the-facts-about-hawaiian-pizza/
   Summary: Not only isHawaiianpizzanot actually from Hawaii, it's not even from the United States. Sotirios (Sam) Panopoulos immigrated to Canada from Greece in 1954 and opened the Satellite Restaurant with his brother in London, Ontario.When the Satellite hired a Chinese-Canadian cook and started adding some sweet and savory dishes to the menu, the inspiration for pineapple onpizzawas born.

3. What Is on a Hawaiian Pizza? Exploring Classic Toppings
   URL: https://recipestasteful.com/what-is-on-a-hawaiian-pizza/
   Summary: Mozzarella Cheese: A creamy, melt-in-your-mouth cheese that blankets the sauce and binds the toppings together. Ham: Often referred to as Canadian bacon, especially in the U.S., this meat adds a savory and slightly smoky flavor. Pineapple: Fresh or canned chunks or slices of pineapple provide a sweet and tangy contrast to the savory ingredients, making it the signature topping of aHawaiianpizza.
```

You see, nothing too magical.

The complete code for this example is available here: [https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/12-use-mcp-toolkit/main.go](https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/12-use-mcp-toolkit/main.go)

It's time now to see how to use this with an LLM.

## Making the LLM recognize "the will to call a tool"

We will apply the concept of **"function calling"** (seen previously: [Function Calling with Docker Model Runner](https://k33g.hashnode.dev/function-calling-with-docker-model-runner)).

The principle of the new example is as follows:

1. Initialize a Docker Model Runner client to connect to the LLM
2. Initialize an MCP client to connect to the MCP server
3. Get the list of "tools" offered by the MCP server
4. Transform the list of "tools" into a format readable by the OpenAI Go SDK API (Docker Model Runner uses the same API)
5. Provide this new list to the LLM to make it detect the call(s) to tool(s) from a prompt (e.g., `"Search information about hawaiian pizza."`)

First, here's the function to convert MCP "tools" to OpenAI "tools":

```golang
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
```

The complete source code is as follows:
```golang
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
    openai.UserMessage("Search information about hawaiian pizza.(only 3 results)"),
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
}
```

To run the example, use the following command:
```bash
MODEL_RUNNER_BASE_URL=http://localhost:12434  MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest go run main.go
```

Or if like me you work with devcontainer:
```bash
MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest go run main.go
```

And you'll get:
```raw
🛠️  Available Tools (OpenAI format):
🔧 Tool: search
  - description: 
    Search DuckDuckGo and return formatted results.

    Args:
        query: The search query string
        max_results: Maximum number of results to return (default: 10)
        ctx: MCP context for logging
    
  - parameters: map[properties:map[max_results:map[default:10 title:Max Results type:integer] query:map[title:Query type:string]] required:[query] type:object]
🔧 Tool: fetch_content
  - description: 
    Fetch and parse content from a webpage URL.

    Args:
        url: The webpage URL to fetch content from
        ctx: MCP context for logging
    
  - parameters: map[properties:map[url:map[title:Url type:string]] required:[url] type:object]

🎉 Detected calls:
📣 calling  search {"query":"hawaiian pizza","max_results":3}
```

🎉 So our LLM is indeed able to understand that we want to call the `search` tool to look for information about `"hawaiian pizza"` with a maximum of `3` results.

The complete code for this example is available here: [https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/13-use-mcp-toolkit/main.go](https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/13-use-mcp-toolkit/main.go)

Now, we need to ask the MCP server to execute the call to the `search` tool.

## Executing tool calls

To ask the MCP server to execute the tool, simply modify the end of the code as follows:

1. For each call, transform the JSON string of arguments into `map[string]any`
2. Use the `mcpClient.CallTool(ctx, toolCall.Function.Name, args)` method
3. Display the results

```golang
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
```

Run the code again with the following command:
```bash
MODEL_RUNNER_BASE_URL=http://localhost:12434  MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest go run main.go
```

Or if you work with devcontainer:
```bash
MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest go run main.go
```

You'll get:
```raw
🎉 Detected calls:
📣 calling  search {"query":"hawaiian pizza","max_results":3}
📝 Arguments: map[max_results:3 query:hawaiian pizza]
🎉📝 Tool response: Found 3 search results:

1. Hawaiian pizza - Wikipedia
   URL: https://en.wikipedia.org/wiki/Hawaiian_pizza
   Summary: Learn about the history and global impact ofHawaiianpizza, a Canadian invention with pineapple, tomato sauce, cheese, and ham or bacon. Find out how people around the world react to this polarizing topping and what celebrities have to say about it.

2. The Best Hawaiian Pizza Recipe | The Recipe Critic
   URL: https://therecipecritic.com/hawaiian-pizza/
   Summary: Learn how to make homemadeHawaiianpizzawith mozzarella cheese,pizzasauce, pineapple tidbits, and Canadian bacon. This sweet and savorypizzais easy, delicious, and perfect forpizzanight.

3. Hawaiian Pizza - Allrecipes
   URL: https://www.allrecipes.com/recipe/8527294/hawaiian-pizza/
   Summary: Learn how to make a crispy and flavorfulHawaiianpizzawithpizzadough,pizzasauce, mozzarella cheese, red onion, Canadian bacon, and pineapple. Follow the easy steps and tips from Allrecipes Test Kitchen staff and enjoy this classicpizzain 45 minutes.
```

If you modify the user prompt:
```golang
openai.UserMessage("Search information about hawaiian pizza.(only 3 results)")
```
to:
```golang
openai.UserMessage(`Search information about hawaiian pizza.(only 3 results)
    Then search information about bananas pizza.(only 1 result)
`),
```

By rerunning the code, you'll get 2 calls to the `search` tool:
```raw
🎉 Detected calls:
📣 calling  search {"query":"hawaiian pizza","max_results":3}
📝 Arguments: map[max_results:3 query:hawaiian pizza]
🎉📝 Tool response: Found 3 search results:

1. The Best Hawaiian Pizza Recipe | The Recipe Critic
   URL: https://therecipecritic.com/hawaiian-pizza/
   Summary: Learn how to make homemadeHawaiianpizzawith mozzarella cheese,pizzasauce, pineapple tidbits, and Canadian bacon. This sweet and savorypizzais easy, delicious, and perfect forpizzanight.

2. Hawaiian Pizza - Allrecipes
   URL: https://www.allrecipes.com/recipe/8527294/hawaiian-pizza/
   Summary: Learn how to make a crispy and flavorfulHawaiianpizzawithpizzadough,pizzasauce, mozzarella cheese, red onion, Canadian bacon, and pineapple. Follow the easy steps and tips from Allrecipes Test Kitchen staff and enjoy this classicpizzain 45 minutes.

3. Hawaiian pizza - Wikipedia
   URL: https://en.wikipedia.org/wiki/Hawaiian_pizza
   Summary: Learn about the history and global impact ofHawaiianpizza, a Canadian invention with pineapple, tomato sauce, cheese, and ham or bacon. Find out how people around the world react to this polarizing topping and what celebrities have to say about it.

📣 calling  search {"query":"bananas pizza","max_results":1}
📝 Arguments: map[max_results:1 query:bananas pizza]
🎉📝 Tool response: Found 1 search results:

1. Banana Cream Pizza - Rhodes Bake-N-Serv
   URL: https://rhodesbakenserv.com/banana-cream-pizza/
   Summary: Place on a 12-inch sprayedpizzapan. Turn up edges of dough to form a ridge. Let rise 10 minutes. Combine brown sugar, butter, and pecans. Sprinkle evenly over crust. Bake at 400ºF 15 minutes. Watch for air bubbles and poke down if needed. Remove from oven. Cool crust completely. Arrangebananason crust. Spoon prepared pudding overbananas.
```

And there you have it! Not so complicated after all, thanks to the **[mcp-golang](https://mcpgolang.com/introduction)** project, **Docker MCP Toolkit**, and **Docker Model Runner** 🥰. In a future blog post, we'll see how to integrate this principle into a complete conversational flow with the LLM (e.g., `"Search information about hawaiian pizza and add fancy and appropriate emojiis to every result"`).

The complete code for this example is available here: [https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/14-use-mcp-toolkit/main.go](https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/14-use-mcp-toolkit/main.go)

Last but not least, how to dockerize this example?

## Dockerizing the example

### Solution 1

In the Go code, we start the MCP process with the following code:

```golang
cmd := exec.Command(
    "docker",
    "run",
    "-i",
    "--rm",
    "alpine/socat",
    "STDIO",
    "TCP:host.docker.internal:8811",
)
```

Which means I'll need `docker` in my container. So instead of using a `scratch` image for my second stage, I'll use the `docker:cli` image.

Here's my `Dockerfile`:
```Dockerfile
FROM golang:1.24.2-alpine AS builder

WORKDIR /app
COPY main.go .
COPY go.mod .

RUN <<EOF
go mod tidy 
go build -o use-mcp-toolkit
EOF

FROM docker:cli
WORKDIR /app
COPY --from=builder /app/use-mcp-toolkit .

CMD ["./use-mcp-toolkit"]
```

Then my `compose.yml` file will be as follows:

```yaml
services:
  use-mcp-toolkit:
    build: .
    environment:
      - MODEL_RUNNER_BASE_URL=${MODEL_RUNNER_BASE_URL}
      - MODEL_RUNNER_LLM_CHAT=${MODEL_RUNNER_LLM_CHAT}
    depends_on:
      - llm-chat
    volumes: 
      - /var/run/docker.sock:/var/run/docker.sock 

  # Download local Docker Model Runner LLMs
  llm-chat:
    provider:
      type: model
      options:
        model: ${MODEL_RUNNER_LLM_CHAT}
```
> Mounting the Docker socket `/var/run/docker.sock` allows the container to communicate with the host's Docker daemon.

If you work with devcontainer, you won't have access to the Compose provider, so you can use this to download the LLM:

```yaml
  # Download local Docker Model Runner LLMs
  llm-chat:
    image: curlimages/curl:8.12.1
    environment:
      - MODEL_RUNNER_BASE_URL=${MODEL_RUNNER_BASE_URL}
      - MODEL_RUNNER_LLM_CHAT=${MODEL_RUNNER_LLM_CHAT}
    entrypoint: |
      sh -c '
      # Download Chat model
      curl -s "${MODEL_RUNNER_BASE_URL}/models/create" -d @- << EOF
      {"from": "${MODEL_RUNNER_LLM_CHAT}"}
      EOF
      '
```
And finally an `.env` file:
```bash
MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal
MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest
```

And now you can launch the dockerized application like this:
```bash
docker compose up --build --no-log-prefix
```

### Solution 2
> Thanks to [@rumpl](https://x.com/rumpl) for help with this part. 🤗

If you don't want to do Docker in Docker, modify the Go code this way:

```golang
// Start the MCP server process
cmd := exec.Command(
    "socat",
    "STDIO",
    "TCP:host.docker.internal:8811",
)
```
> Socat (Socket CAT) is a command-line utility for Unix systems that allows establishing bidirectional connections between different types of communication channels.

Then, modify the `Dockerfile` like this:
```Dockerfile
FROM golang:1.24.2-alpine AS builder

WORKDIR /app
COPY main.go .
COPY go.mod .

RUN <<EOF
go mod tidy
go build -o use-mcp-toolkit
EOF

# Final stage
FROM alpine:3.19
WORKDIR /app

# Install socat
RUN apk add --no-cache socat

COPY --from=builder /app/use-mcp-toolkit .

CMD ["./use-mcp-toolkit"]
```

And finally the `compose.yml` file like this:
```yaml
# docker compose up --build --no-log-prefix
services:
  use-mcp-toolkit:
    build: .
    environment:
      - MODEL_RUNNER_BASE_URL=${MODEL_RUNNER_BASE_URL}
      - MODEL_RUNNER_LLM_CHAT=${MODEL_RUNNER_LLM_CHAT}
    depends_on:
      - llm-chat

  # Download local Docker Model Runner LLMs
  llm-chat:
    image: curlimages/curl:8.12.1
    environment:
      - MODEL_RUNNER_BASE_URL=${MODEL_RUNNER_BASE_URL}
      - MODEL_RUNNER_LLM_CHAT=${MODEL_RUNNER_LLM_CHAT}
    entrypoint: |
      sh -c '
      # Download Chat model
      curl -s "${MODEL_RUNNER_BASE_URL}/models/create" -d @- << EOF
      {"from": "${MODEL_RUNNER_LLM_CHAT}"}
      EOF
      '
  # Or like this:

  #llm-chat:
  #  provider:
  #    type: model
  #    options:
  #      model: ${MODEL_RUNNER_LLM_CHAT}
```

And now relaunch the application:
```bash
docker compose up --build --no-log-prefix
```

You'll find the complete code for this example here: [https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/15-use-mcp-toolkit/main.go](https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/15-use-mcp-toolkit/main.go)

That's it for this blog post. See you soon for the continuation.
