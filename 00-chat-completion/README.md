# First Chat completion with Model Runner

```bash
docker build -t quick-chat .
docker run -e MODEL_RUNNER_BASE_URL=http://model-runner.docker.internal quick-chat
```