# Generating Structured Data with Docker Model Runner

When I'm programming, I often need realistic test data. For this purpose, there are frameworks like [Faker](https://github.com/faker-js/faker), [Java Faker](https://github.com/DiUS/java-faker), and others. Today, thanks to the "JSON structured output" principle that allows constraining a model to generate data according to a structured data schema, you can generate your own "fake" data with an LLM. Note that:
- The data won't be so "fake" since it comes from the model's training data
- Generation can be time-consuming depending on the model size and your machine's power.

Let's see how to proceed. Once again, we'll use the OpenAI Golang SDK with Docker Model Runner.

## Program and Model Initialization

The `main` function will start as follows:
```golang
ctx := context.Background()

// Docker Model Runner base URL
chatURL := os.Getenv("MODEL_RUNNER_BASE_URL") + "/engines/llama.cpp/v1/"
model := os.Getenv("MODEL_RUNNER_LLM_CHAT")

client := openai.NewClient(
    option.WithBaseURL(chatURL),
    option.WithAPIKey(""),
)
```
> For this example, I'm using the following values:
> - MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest
> - MODEL_RUNNER_BASE_URL=http://localhost:12434
> Or if you're running the code from a container (for example, if you're using devcontainer):
> - MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal


## The Data Schema

Let's say I want to generate data related to a country in the following format:

```json
{
  "capital": "Ottawa",
  "languages": ["English", "French"],
  "name": "Canada"
}
```

For this, I'll need to describe a schema that I'll then use with the SDK to have the LLM generate data in the right format:

```golang
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
```

1. `schema := map[string]any{...}` - Creates a map where keys are strings and values can be of any type.
2. `"type": "object"` - Defines that this schema represents a JSON object.
3. `"properties": map[string]any{...}` - Defines the properties this object can contain.
4. The defined properties are:
   - `"name"`: a string
   - `"capital"`: a string
   - `"languages"`: an array of strings
5. `"required": []string{"name", "capital", "languages"}` - Specifies that all three properties are required.

Once the schema is defined, I'll use it to create a parameter for the request to send to the LLM:

```golang
schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
    Name:        "country_info",
    Description: openai.String("Notable information about a country in the world"),
    Schema:      schema,
    Strict:      openai.Bool(true),
}
```

## Data Generation

We now have everything we need to generate our data:

```golang
userQuestion := openai.UserMessage(`
    Tell me about Canada.
`)

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
    panic(err)
}
fmt.Println("Response:", completion.Choices[0].Message.Content)
```
> ✋ note: the request parameter `ResponseFormat` which allows passing `schemaParam` to the LLM.

If I run the code, 

```bash
MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal \
MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest \
go run main.go
```

I'll get:

```raw
Response: {
  "capital": "Ottawa",
  "languages": ["English", "French"],
  "name": "Canada"
}
```
> You can try with other countries to verify.

The complete code for this example is available here: [https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/08-structured-output/main.go](https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/08-structured-output/main.go)

But let's go a bit further.

## "Give me a list of 10 countries"

We can also ask the LLM to give us a certain number of countries on a given continent with the following schema:

```golang
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
```

And this time my question will be:

```golang
userQuestion := openai.UserMessage("List of 10 countries in Europe")
```

The rest of the code is similar to the previous example:

```golang
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
    panic(err)
}

data := completion.Choices[0].Message.Content

fmt.Println("Countries List:\n", data)
```

If I run the code, 

```bash
MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal \
MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest \
go run main.go
```

I'll get:

```raw
Countries List: 
{ 
    "countries": 
        [ 
            "Albania", 
            "Andorra", 
            "Austria", 
            "Belarus", 
            "Belgium", 
            "Bosnia and Herzegovina", 
            "Bulgaria", 
            "Croatia", 
            "Cyprus", 
            "Czech Republic" 
        ] 
}
```

The complete code for this example is available here: [https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/09-structured-output/main.go](https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/09-structured-output/main.go)

So now, if I wanted to get the information from the first example (capital, languages) for each of these countries, I could iterate through the list of countries and make a request with the previous schema for each one to get the information:

```golang
var countriesList map[string][]string

err = json.Unmarshal([]byte(data), &countriesList)

if err != nil {
    panic(err)
}

fmt.Println("Countries List:")

schema = map[string]any{
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

schemaParam = openai.ResponseFormatJSONSchemaJSONSchemaParam{
    Name:        "country_info",
    Description: openai.String("Notable information about a country in the world"),
    Schema:      schema,
    Strict:      openai.Bool(true),
}

for idx, country := range countriesList["countries"] {
    fmt.Println(idx, ".", country)
    userQuestion := openai.UserMessage("Tell me about " + country)

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
        panic(err)
    }
    fmt.Println("Response:", completion.Choices[0].Message.Content)

}
```

The complete code for this example is available here: [https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/10-structured-output/main.go](https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/10-structured-output/main.go)

But actually, there's a simpler way 🙂

## "Give me a list of 10 countries and their information"

In fact, we can use a single schema:

```golang
schema := map[string]any{
    "type": "object",
    "properties": map[string]any{
        "countries": map[string]any{
            "type": "array",
            "items": map[string]any{
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
            },
        },
    },
    "required": []string{"countries"},
}
```

So this time the rest of the code will be simpler:

```golang
schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
    Name:        "List of countries",
    Description: openai.String("List of countries in the world"),
    Schema:      schema,
    Strict:      openai.Bool(true),
}

userQuestion := openai.UserMessage("List of 10 countries in Europe")

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
    panic(err)
}

data := completion.Choices[0].Message.Content

var countriesList map[string][]any

err = json.Unmarshal([]byte(data), &countriesList)

if err != nil {
    panic(err)
}
fmt.Println("Countries List:")
for idx, country := range countriesList["countries"] {
    fmt.Println(idx, ".", country)
}
```

If I run the code, 

```bash
MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal \
MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest \
go run main.go
```

I'll get:

```raw
Countries List:
0 . map[capital:Athens languages:[Greek] name:Greece]
1 . map[capital:Vienna languages:[German] name:Austria]
2 . map[capital:Berlin languages:[German] name:Germany]
3 . map[capital:Prague languages:[Czech] name:Czech Republic]
4 . map[capital:Budapest languages:[Hungarian] name:Hungary]
5 . map[capital:Rome languages:[Italian] name:Italy]
6 . map[capital:Paris languages:[French] name:France]
7 . map[capital:Lisbon languages:[Portuguese] name:Portugal]
8 . map[capital:Madrid languages:[Spanish] name:Spain]
9 . map[capital:Stockholm languages:[Swedish] name:Sweden]
```

🎉, there you go, pretty easy and convenient, right?

The complete code for this example is available here: [https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/11-structured-output/main.go](https://github.com/Short-Compendium/docker-model-runner-with-golang/blob/main/11-structured-output/main.go)

Before leaving, let's see how to run our application with Docker Compose, which offers Docker Model Runner integration.

## Starting the Application with Docker Compose

Since the very recent version of Docker Desktop (`4.41`), you can ask Docker Compose to handle downloading the LLM(s) necessary for your application (if the LLM is not already present locally), thanks to **[provider services](https://docs.docker.com/compose/how-tos/model-runner/#provider-services)**:

```yaml
llm-chat:
  provider:
    type: model
    options:
      model: ${MODEL_RUNNER_LLM_CHAT}
```

So to "dockerize", I'll create a `Dockerfile`:

```Dockerfile
FROM golang:1.24.2-alpine AS builder

WORKDIR /app
COPY main.go .
COPY go.mod .

RUN <<EOF
go mod tidy 
go build -o structured-output
EOF

FROM scratch
WORKDIR /app
COPY --from=builder /app/structured-output .

CMD ["./structured-output"]
```

Then a `.env` file:
```bash
MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal
MODEL_RUNNER_LLM_CHAT=ai/qwen2.5:latest
```

and finally a `compose.yml` file:
```yaml
services:
  structured-output:
    build: .
    environment:
      - MODEL_RUNNER_BASE_URL=${MODEL_RUNNER_BASE_URL}
      - MODEL_RUNNER_LLM_CHAT=${MODEL_RUNNER_LLM_CHAT}
    depends_on:
      - llm-chat

  # Download local Docker Model Runner LLMs
  
  llm-chat:
    provider:
      type: model
      options:
        model: ${MODEL_RUNNER_LLM_CHAT}
```

And you can launch the application as follows:
```bash
docker compose up --build --no-log-prefix
```

✋ **If you're using Dev Container in VSCode**, your application runs on Linux, and currently **Docker Model Runner** is not implemented on Linux. Nevertheless, you have access to the **Docker Model Runner** REST API served by **Docker Desktop**, so your Go application will work, but not the Docker Compose **[model runner provider service](https://docs.docker.com/compose/how-tos/model-runner/#provider-services)**. You'll need to use this instead:

```yaml
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

## Conclusion

Once again, we've seen that a local LLM can provide great services and that it's not always necessary to call upon its bigger siblings (OpenAI, Claude.ai, Gemini, etc.).

See you soon for a dive into the field of Model Context Protocol Servers with Docker and Docker Model Runner. Stay tuned. 👋