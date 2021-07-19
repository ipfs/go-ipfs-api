package shell

import (
	"fmt"
	"testing"
	"time"
)

var (
	//lint:ignore U1000 used only by skipped tests at present
	examplesHashForIPNS = "/ipfs/Qmbu7x6gJbsKDcseQv66pSbUcAA3Au6f7MfTYVXwvBxN2K"
	//lint:ignore U1000 used only by skipped tests at present
	testKey = "self" // feel free to change to whatever key you have locally
)

func TestPublishDetailsWithKey(t *testing.T) {
	t.Skip()
	shell := NewShell("localhost:5001")

	resp, err := shell.PublishWithDetails(examplesHashForIPNS, testKey, time.Second, time.Second, false)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Value != examplesHashForIPNS {
		t.Fatalf(fmt.Sprintf("Expected to receive %s but got %s", examplesHash, resp.Value))
	}
}

func TestPublishDetailsWithoutKey(t *testing.T) {
	t.Skip()
	shell := NewShell("localhost:5001")

	resp, err := shell.PublishWithDetails(examplesHashForIPNS, "", time.Second, time.Second, false)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Value != examplesHashForIPNS {
		t.Fatalf(fmt.Sprintf("Expected to receive %s but got %s", examplesHash, resp.Value))
	}
}
