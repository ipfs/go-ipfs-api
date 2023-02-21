package shell

import (
	"context"
	"os"
	"testing"

	"github.com/cheekybits/is"
)

func TestKeyGen(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	defer func() {
		_, err := s.KeyRm(context.Background(), "testKey1")
		is.Nil(err)
	}()
	key1, err := s.KeyGen(context.Background(), "testKey1", KeyGen.Type("ed25519"))
	is.Nil(err)
	is.Equal(key1.Name, "testKey1")
	is.NotNil(key1.Id)

	defer func() {
		_, err = s.KeyRm(context.Background(), "testKey2")
		is.Nil(err)
	}()
	key2, err := s.KeyGen(context.Background(), "testKey2", KeyGen.Type("ed25519"))
	is.Nil(err)
	is.Equal(key2.Name, "testKey2")
	is.NotNil(key2.Id)

	defer func() {
		_, err = s.KeyRm(context.Background(), "testKey3")
		is.Nil(err)
	}()
	key3, err := s.KeyGen(context.Background(), "testKey3", KeyGen.Type("rsa"))
	is.Nil(err)
	is.Equal(key3.Name, "testKey3")
	is.NotNil(key3.Id)

	defer func() {
		_, err = s.KeyRm(context.Background(), "testKey4")
		is.Nil(err)
	}()
	key4, err := s.KeyGen(context.Background(), "testKey4", KeyGen.Type("rsa"), KeyGen.Size(4096))
	is.Nil(err)
	is.Equal(key4.Name, "testKey4")
	is.NotNil(key4.Id)

	_, err = s.KeyGen(context.Background(), "testKey5", KeyGen.Type("rsa"), KeyGen.Size(1024))
	is.NotNil(err)
	is.Equal(err.Error(), "key/gen: rsa keys must be >= 2048 bits to be useful")
}

func TestKeyList(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	defer func() {
		_, err := s.KeyRm(context.Background(), "testKey")
		is.Nil(err)
	}()
	key, err := s.KeyGen(context.Background(), "testKey")
	is.Nil(err)

	keys, err := s.KeyList(context.Background())
	is.Nil(err)

	is.Equal(len(keys), 2)
	is.Equal(keys[0].Name, "self")
	is.NotNil(keys[0].Id)
	is.NotNil(keys[1].Id, key.Id)
	is.NotNil(keys[1].Name, key.Name)
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

	_, err = s.KeyRm(context.Background(), "test2")
	is.Nil(err)

	_, err = s.KeyRename(context.Background(), "test2", "test1", false)
	is.NotNil(err)
	is.Equal(err.Error(), "key/rename: no key named test2 was found")
}

func TestKeyRm(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	key, err := s.KeyGen(context.Background(), "testKey")
	is.Nil(err)

	keys, err := s.KeyRm(context.Background(), "testKey")
	is.Nil(err)

	is.Equal(len(keys), 1)
	is.Equal(keys[0].Name, key.Name)
	is.Equal(keys[0].Id, key.Id)

	_, err = s.KeyRm(context.Background(), "testKey")
	is.NotNil(err)
	is.Equal(err.Error(), "key/rm: no key named testKey was found")
}

func TestKeyImport(t *testing.T) {
	is := is.New(t)
	s := NewShell(shellUrl)

	// Key generated as per: https://github.com/ipfs/kubo/blob/c9d51bbe0133968858aa9991b7f69ec269126599/test/sharness/t0165-keystore-data/README.md
	f, err := os.Open("./tests/key_test.pem")
	is.Nil(err)
	defer f.Close()

	err = s.KeyImport(context.Background(), "testImportKey", f, KeyImportGen.Format("pem-pkcs8-cleartext"))
	is.Nil(err)

	_, err = s.KeyRm(context.Background(), "testImportKey")
	is.Nil(err)
}
