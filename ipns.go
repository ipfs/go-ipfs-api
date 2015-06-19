package shell

import (
	cmds "github.com/ipfs/go-ipfs/commands"
	cc "github.com/ipfs/go-ipfs/core/commands"
)

// Publish updates a mutable name to point to a given value
func (s *Shell) Publish(node string, value string) error {
	ropts, err := cc.Root.GetOptions([]string{"name", "publish"})
	if err != nil {
		return err
	}

	args := []string{value}
	if node != "" {
		args = append([]string{node}, args...)
	}
	req, err := cmds.NewRequest([]string{"name", "publish"}, nil, args, nil, cc.PublishCmd, ropts)
	if err != nil {
		return err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return err
	}
	if resp.Error() != nil {
		return resp.Error()
	}

	return nil
}
