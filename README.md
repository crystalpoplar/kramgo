# kramgo

A cross-platform CLI tool for AI interactions, written in Go.

## Features

- Chat with AI providers (OpenAI, Anthropic) directly from your terminal
- Interactive chat sessions
- Simple configuration management
- Cross-platform: Linux, macOS, Windows (amd64 and arm64)

## Installation

```sh
go install github.com/crystalpoplar/kramgo/cmd/kramgo@latest
```

Or build from source:

```sh
make build
# binary is at bin/kramgo
```

## Usage

```sh
# Send a single message
kramgo chat --message "What is the capital of France?"

# Interactive session
kramgo chat --interactive

# Show current config
kramgo config show

# Set your API key
kramgo config set api_key sk-...

# Switch provider
kramgo config set provider anthropic
kramgo config set model claude-3-5-sonnet-20241022

# Print version
kramgo version
```

## Configuration

Configuration is stored in:

- Linux / macOS: `~/.config/kramgo/config.json`
- Windows: `%APPDATA%\kramgo\config.json`

You can also provide the API key via environment variable:

| Provider  | Environment variable |
|-----------|---------------------|
| openai    | `OPENAI_API_KEY`    |
| anthropic | `ANTHROPIC_API_KEY` |

## Development

```sh
make test   # run tests
make lint   # run go vet
make cross  # cross-compile for all supported platforms
```