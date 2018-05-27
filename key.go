package shell

import (
	"context"
	"errors"
	"encoding/json"

	"github.com/ipfs/go-ipfs/core/coreapi/interface"
	"github.com/ipfs/go-ipfs/core/coreapi/interface/options"
)

type httpKeyApi httpApi

type key struct {
	KeyName string `json:"Name"`
	Id      string `json:"Id"`
}

func (k *key) Name() string {
	return k.KeyName
}

func (k *key) Path() iface.Path {
	return nil //TODO
}


func (api *httpKeyApi) Generate(ctx context.Context, name string, opts ...options.KeyGenerateOption) (iface.Key, error) {
	return nil, errors.New("TODO")
}

func (api *httpKeyApi) WithType(algorithm string) options.KeyGenerateOption {
	return nil
}

func (api *httpKeyApi) WithSize(size int) options.KeyGenerateOption {
	return nil
}

func (api *httpKeyApi) Rename(ctx context.Context, oldName string, newName string, opts ...options.KeyRenameOption) (iface.Key, bool, error) {
	return nil, false, errors.New("TODO")
}

func (api *httpKeyApi) WithForce(force bool) options.KeyRenameOption {
	return nil
}

func (api *httpKeyApi) List(ctx context.Context) ([]iface.Key, error) {
	resp, err := api.core().newRequest(ctx, "key/list").Send(api.client)
	if err != nil {
		return nil, err
	}

	var res = struct{
		Keys []*key
	}{}

	if err := json.NewDecoder(resp.Output).Decode(&res); err != nil {
		return nil, err
	}

	out := make([]iface.Key, len(res.Keys))
	for i, e := range res.Keys {
		out[i] = e
	}

	return out, nil
}

func (api *httpKeyApi) Remove(ctx context.Context, name string) (iface.Path, error) {
	return nil, errors.New("TODO")
}

func (api *httpKeyApi) core() *httpApi {
	return (*httpApi)(api)
}
