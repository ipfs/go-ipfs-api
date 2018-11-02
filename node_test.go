package shell

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TODO add more unit tests

const (
	ipfsTestRoot string = "test"
)

func TestNode_Init(t *testing.T) {
	as := assert.New(t)
	defer os.RemoveAll(ipfsTestRoot)

	cfg := NodeConfig{
		Root:       ipfsTestRoot,
		StorageMax: "10G",
	}
	node := NewNode(&cfg)

	err := node.Init()
	as.Nil(err)

	err2 := node.Init()
	as.NotNil(err2)
}
