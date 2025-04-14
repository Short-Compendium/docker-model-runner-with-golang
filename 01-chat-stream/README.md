# First Chat Stream completion with Model Runner

```bash
docker build -t quick-chat-stream .
docker run -e MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal quick-chat-stream
```