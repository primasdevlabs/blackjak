package llm

import (
	"context"
)

// ProviderClient adapts a Provider to the Client interface by pinning a
// default model. The agent runtime depends on Client; the settings layer
// resolves which Provider/model pair to use per role.
type ProviderClient struct {
	Provider Provider
	Model    string
}

// NewProviderClient creates a Client backed by provider using defaultModel.
func NewProviderClient(provider Provider, defaultModel string) *ProviderClient {
	return &ProviderClient{Provider: provider, Model: defaultModel}
}

// Complete routes to Provider.Chat, filling in the model when unspecified.
func (c *ProviderClient) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	if req.Model == "" {
		req.Model = c.Model
	}
	return c.Provider.Chat(ctx, *req)
}
