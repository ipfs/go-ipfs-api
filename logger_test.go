package shell

import "testing"

func Test_Logger(t *testing.T) {
	sh := NewLocalShell()
	logger, err := sh.GetLogs()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := logger.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	_, err = logger.Next()
	if err != nil {
		t.Fatal(err)
	}
}
