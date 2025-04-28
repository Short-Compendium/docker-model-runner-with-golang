#!/bin/sh
set -e

curl -s "${MODEL_RUNNER_BASE_URL}/models/create" -d '{"from": "ai/mistral:latest"}'
curl -s "${MODEL_RUNNER_BASE_URL}/models/create" -d '{"from": "ai/llama3.2"}'
curl -s "${MODEL_RUNNER_BASE_URL}/models/create" -d '{"from": "ai/mxbai-embed-large"}'
curl -s "${MODEL_RUNNER_BASE_URL}/models/create" -d '{"from": "ai/qwen2.5:latest"}'
curl -s "${MODEL_RUNNER_BASE_URL}/models/create" -d '{"from": "ai/smollm2"}'
