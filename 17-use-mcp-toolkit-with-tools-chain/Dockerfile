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