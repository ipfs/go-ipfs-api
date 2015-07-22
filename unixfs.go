package shell

import (
	"encoding/json"
	"fmt"

	cmds "github.com/ipfs/go-ipfs/commands"
	cc "github.com/ipfs/go-ipfs/core/commands"
	unixfs "github.com/ipfs/go-ipfs/core/commands/unixfs"
)

// FileList entries at the given path using the UnixFS commands
func (s *Shell) FileList(path string) (*unixfs.LsObject, error) {
	ropts, err := cc.Root.GetOptions([]string{"file", "ls"})
	if err != nil {
		return nil, err
	}

	req, err := cmds.NewRequest([]string{"file", "ls"}, nil, []string{path}, nil, unixfs.LsCmd, ropts)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	if resp.Error() != nil {
		return nil, resp.Error()
	}

	read, err := resp.Reader()
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(read)
	out := unixfs.LsOutput{}
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}

	for _, object := range out.Objects {
		return object, nil
	}

	return nil, fmt.Errorf("no object in results")
}
