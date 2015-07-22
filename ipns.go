package shell

import (
	"encoding/json"

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
	defer resp.Close()
	if resp.Error() != nil {
		return resp.Error()
	}

	return nil
}

func (s *Shell) Resolve(id string) (string, error) {
	ropts, err := cc.Root.GetOptions([]string{"name", "resolve"})
	if err != nil {
		return "", err
	}

	req, err := cmds.NewRequest([]string{"name", "resolve"}, nil, []string{id}, nil, cc.ResolveCmd, ropts)
	if err != nil {
		return "", err
	}

	resp, err := s.client.Send(req)
	if err != nil {
		return "", err
	}
	defer resp.Close()
	if resp.Error() != nil {
		return "", resp.Error()
	}

	r, err := resp.Reader()
	if err != nil {
		return "", err
	}

	out := struct {
		Path string
	}{}

	err = json.NewDecoder(r).Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Path, nil
}
