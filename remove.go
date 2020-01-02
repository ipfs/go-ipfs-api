package shell

import (
	"context"
	"strings"
)

// Remove file by hash.
func (s *Shell) Remove(hash string) bool {
	var out string
	rb := s.Request("rm", hash)
	if err := rb.Exec(context.Background(), &out); err != nil {
		return false
	}

	if strings.Contains(out, "Removed") {
		return true
	}

	return false
}
