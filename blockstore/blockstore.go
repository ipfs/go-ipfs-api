package blockstore

import (
	"context"
	"fmt"

	"github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/go-ipfs-blockstore"
	mh "github.com/multiformats/go-multihash"
)

// ApiBlockstore implements the go-ipfs-blockstore interface
// using an ipfs node's http api as the backend
type ApiBlockstore struct {
	Shell *shell.Shell
}

var _ blockstore.Blockstore = (*ApiBlockstore)(nil)

func (bs *ApiBlockstore) AllKeysChan(ctx context.Context) (<-chan *cid.Cid, error) {
	// TODO: we could use 'ipfs refs local' here
	return nil, fmt.Errorf("unsupported")
}

func (bs *ApiBlockstore) DeleteBlock(c *cid.Cid) error {
	return fmt.Errorf("delete is unsupported via the api")
}

func (bs *ApiBlockstore) Get(c *cid.Cid) (blocks.Block, error) {
	data, err := bs.Shell.BlockGet(c.String())
	if err != nil {
		return nil, err
	}

	return blocks.NewBlockWithCid(data, c)
}

func (bs *ApiBlockstore) Has(c *cid.Cid) (bool, error) {
	// TODO: check for a 'not found' error
	_, _, err := bs.Shell.BlockStat(c.String())
	return err == nil, nil
}

func (bs *ApiBlockstore) Put(b blocks.Block) error {
	pref := b.Cid().Prefix()

	format, ok := cid.CodecToStr[pref.Codec]
	if !ok {
		return fmt.Errorf("unsupported codec on block cid")
	}

	mhtyp, ok := mh.Codes[pref.MhType]
	if !ok {
		return fmt.Errorf("multihash in cid had unknown type")
	}

	// TODO: maybe check the cid we got back was right?
	_, err := bs.Shell.BlockPut(b.RawData(), format, mhtyp, pref.MhLength)
	if err != nil {
		return err
	}

	return nil
}

func (bs *ApiBlockstore) PutMany(blks []blocks.Block) error {
	for _, blk := range blks {
		if err := bs.Put(blk); err != nil {
			return err
		}
	}
	return nil
}

func (bs *ApiBlockstore) HashOnRead(_ bool) {
	// this technically always happens
}
