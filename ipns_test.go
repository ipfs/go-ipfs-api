package shell_test

import (
	"fmt"
	"testing"
	"time"

	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
)

var examplesHash = "/ipfs/Qmbu7x6gJbsKDcseQv66pSbUcAA3Au6f7MfTYVXwvBxN2K"

func TestPublishDetailsWithKey(t *testing.T) {
	shell := ipfsapi.NewShell("localhost:5001")

	resp, err := shell.PublishWithDetails(examplesHash, "self", time.Hour, time.Hour, false)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Value != examplesHash {
		t.Fatalf(fmt.Sprintf("Expected to receive %s but got %s", examplesHash, resp.Value))
	}
}

func TestPublishDetailsWithoutKey(t *testing.T) {
	shell := ipfsapi.NewShell("localhost:5001")

	resp, err := shell.PublishWithDetails(examplesHash, "", time.Hour, time.Hour, false)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Value != examplesHash {
		t.Fatalf(fmt.Sprintf("Expected to receive %s but got %s", examplesHash, resp.Value))
	}
}
