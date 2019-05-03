package shell

import (
	"testing"
)

func Test_Logger(t *testing.T) {
	sh := NewShell(shellUrl)
	logger, err := sh.GetLogs()
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
