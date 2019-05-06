package shell

import (
	"context"
	"testing"
)

func TestLogger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sh := NewShell(shellUrl)
	logger, err := sh.GetLogs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	if l, err := logger.Next(); err != nil {
		t.Fatal(err)
	} else if l == nil {
		t.Fatal("no logs found")
	}
}
