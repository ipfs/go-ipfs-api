package shell

import (
	"testing"
	"context"
)

//TODO: create utils to run tests on a proper test node
func TestListSelf(t *testing.T) {
	ctx := context.Background()
	api, err := NewLocalApi()
	if err != nil {
		t.Fatal(err)
	}

	keys, err := api.Key().List(ctx)
	if err != nil {
		t.Fatalf("failed to list keys: %s", err)
	}

	if keys[0].Name() != "self" {
		t.Errorf("expected the key to be called 'self', got '%s'", keys[0].Name())
	}
}
