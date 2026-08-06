// Package ai provides a provider-agnostic interface for AI chat completions.
package ai

import (
	"context"
	"errors"
	"fmt"
)

// Message represents a single chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionRequest holds the parameters for a chat completion.
type CompletionRequest struct {
	Model    string
	Messages []Message
}

// CompletionResponse holds the result of a chat completion.
type CompletionResponse struct {
	Content string
	Model   string
}

// Provider is the interface that all AI backend implementations must satisfy.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Name() string
}

// Client wraps a Provider and exposes a simple chat method.
type Client struct {
	provider Provider
}

// NewClient returns a Client backed by provider.
func NewClient(provider Provider) *Client {
	return &Client{provider: provider}
}

// Chat sends a single user message and returns the assistant's reply.
func (c *Client) Chat(ctx context.Context, model, userMessage string) (string, error) {
	req := CompletionRequest{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: userMessage},
		},
	}
	resp, err := c.provider.Complete(ctx, req)
	if err != nil {
		return "", fmt.Errorf("ai: completion failed: %w", err)
	}
	return resp.Content, nil
}

// ProviderFactory creates a Provider from a name and API key.
func ProviderFactory(name, apiKey string) (Provider, error) {
	switch name {
	case "openai":
		return NewOpenAIProvider(apiKey), nil
	case "anthropic":
		return NewAnthropicProvider(apiKey), nil
	default:
		return nil, errors.New("ai: unknown provider: " + name)
	}
}
