package shell

import (
	"context"
	"testing"

	"github.com/cheekybits/is"
)

func TestKeyList(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	keys, err := s.KeyList(context.Background())
	is.Nil(err)

	is.Equal(len(keys), 1)
	is.Equal(keys[0].Name, "self")
	is.NotNil(keys[0].Id)
}
