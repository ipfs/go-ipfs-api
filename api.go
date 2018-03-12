package shell

import (
	"github.com/ipfs/go-ipfs/core/coreapi/interface"
	"context"
	"errors"
)

type httpApi struct{}

func NewLocalApi() (iface.CoreAPI, error) {
	return &httpApi{}, nil
}

// Unixfs returns the UnixfsAPI interface backed by the go-ipfs node
func (api *httpApi) Unixfs() iface.UnixfsAPI {
	return nil
}

func (api *httpApi) Block() iface.BlockAPI {
	return nil
}

// Dag returns the DagAPI interface backed by the go-ipfs node
func (api *httpApi) Dag() iface.DagAPI {
	return nil
}

// Name returns the NameAPI interface backed by the go-ipfs node
func (api *httpApi) Name() iface.NameAPI {
	return nil
}

// Key returns the KeyAPI interface backed by the go-ipfs node
func (api *httpApi) Key() iface.KeyAPI {
	return nil
}

//Object returns the ObjectAPI interface backed by the go-ipfs node
func (api *httpApi) Object() iface.ObjectAPI {
	return nil
}

func (api *httpApi) Pin() iface.PinAPI {
	return nil
}

// ResolveNode resolves the path `p` using Unixfs resolver, gets and returns the
// resolved Node.
func (api *httpApi) ResolveNode(ctx context.Context, p iface.Path) (iface.Node, error) {
	return nil, errors.New("TODO")
}

// ResolvePath resolves the path `p` using Unixfs resolver, returns the
// resolved path.
func (api *httpApi) ResolvePath(ctx context.Context, p iface.Path) (iface.Path, error) {
	return nil, errors.New("TODO")
}
