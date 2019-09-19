package vault

import (
	"context"
)

// ActiveTokens contains the number of active tokens.
type ActiveTokens struct {
	// ServiceTokens contains information about the number of active service
	// tokens.
	ServiceTokens TokenCounter `json:"service_tokens"`
}

// TokenCounter counts the number of tokens
type TokenCounter struct {
	// Total is the total number of tokens
	Total int `json:"total"`
}

// countActiveTokens returns the number of active tokens
func (c *Core) countActiveTokens(ctx context.Context) (*ActiveTokens, error) {

	// Get all of the namespaces
	ns := c.collectNamespaces()

	// Count the tokens under each namespace
	total := 0
	for i := 0; i < len(ns); i++ {
		ids, err := c.tokenStore.idView(ns[i]).List(ctx, "")
		if err != nil {
			return nil, err
		}
		total += len(ids)
	}

	return &ActiveTokens{
		ServiceTokens: TokenCounter{
			Total: total,
		},
	}, nil
}
