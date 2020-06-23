package shell

import (
	"context"
	"testing"

	"github.com/cheekybits/is"
)

func TestKeyGen(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	key, err := s.KeyGen(context.Background(), "testKey", KeyGen.Type("ed25519"), KeyGen.Size(2048))
	is.Nil(err)

	is.Equal(key.Name, "testKey")
	is.NotNil(key.Id)
}

func TestKeyList(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	keys, err := s.KeyList(context.Background())
	is.Nil(err)

	is.Equal(len(keys), 1)
	is.Equal(keys[0].Name, "self")
	is.NotNil(keys[0].Id)
}

func TestKeyRename(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	key, err := s.KeyGen(context.Background(), "test1")
	is.Nil(err)

	out, err := s.KeyRename(context.Background(), "test1", "test2", false)
	is.Nil(err)

	is.Equal(out.Now, "test2")
	is.Equal(out.Was, "test1")
	is.Equal(out.Id, key.Id)
	is.False(out.Overwrite)
}
