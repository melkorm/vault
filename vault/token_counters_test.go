package vault

import (
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/vault/helper/namespace"
	"github.com/hashicorp/vault/sdk/logical"
)

func testCountActiveTokens(t *testing.T, c *Core, root string, expectedServiceTokens int) {
	rootCtx := namespace.RootContext(nil)
	resp, err := c.HandleRequest(rootCtx, &logical.Request{
		ClientToken: root,
		Operation:   logical.ReadOperation,
		Path:        "sys/internal/counters/tokens",
	})
	if err != nil || (resp != nil && resp.IsError()) {
		t.Fatalf("bad: resp: %#v\n err: %v", resp, err)
	}

	if diff := deep.Equal(resp.Data, map[string]interface{}{
		"counters": &ActiveTokens{
			ServiceTokens: TokenCounter{
				Total: expectedServiceTokens,
			},
		},
	}); diff != nil {
		t.Fatal(diff)
	}
}

func TestTokenStore_CountActiveTokens(t *testing.T) {
	c, _, root := TestCoreUnsealed(t)
	rootCtx := namespace.RootContext(nil)

	// Count the root token
	testCountActiveTokens(t, c, root, 1)

	// Create some service tokens
	req := &logical.Request{
		ClientToken: root,
		Operation:   logical.UpdateOperation,
		Path:        "create",
	}
	tokens := make([]string, 10)
	for i := 0; i < 10; i++ {
		resp, err := c.tokenStore.HandleRequest(rootCtx, req)
		if err != nil || (resp != nil && resp.IsError()) {
			t.Fatalf("bad: resp: %#v\n err: %v", resp, err)
		}
		tokens[i] = resp.Auth.ClientToken

		testCountActiveTokens(t, c, root, i+2)
	}

	// Revoke the service tokens
	req.Path = "revoke"
	req.Data = make(map[string]interface{})
	for i := 0; i < 10; i++ {
		req.Data["token"] = tokens[i]
		resp, err := c.tokenStore.HandleRequest(rootCtx, req)
		if err != nil || (resp != nil && resp.IsError()) {
			t.Fatalf("bad: resp: %#v\n err: %v", resp, err)
		}

		testCountActiveTokens(t, c, root, 10-i)
	}
}
