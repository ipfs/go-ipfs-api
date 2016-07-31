package shell

import (
	"encoding/json"
)

// Publish updates a mutable name to point to a given value
func (s *Shell) Publish(node string, value string) error {
	args := []string{value}
	if node != "" {
		args = []string{node, value}
	}

	resp, err := s.newRequest("name/publish", args...).Send(s.httpcli)
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

// Resolve resolves the string provided to an /ipfs/[hash]. If asked to
// resolve an empty string, resolve instead resolves the node's own /ipns value.
func (s *Shell) Resolve(id string) (string, error) {
	return s.resolve(id, false)
}

// ResolveFresh resolves the string provided to an /ipfs/[hash] without looking
// at the cache. If asked to resolve an empty string, ResolveFresh instead
// resolves the node's own /ipns value.
func (s *Shell) ResolveFresh(id string) (string, error) {
	return s.resolve(id, true)
}

func (s *Shell) resolve(id string, nocache bool) (string, error) {
	var req *Request
	if id != "" {
		req = s.newRequest("name/resolve", id)
	} else {
		req = s.newRequest("name/resolve")
	}

	if nocache {
		req.Opts["nocache"] = "true"
	}
	// false is the default

	resp, err := req.Send(s.httpcli)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	if resp.Error != nil {
		return "", resp.Error
	}

	var out struct{ Path string }
	err = json.NewDecoder(resp.Output).Decode(&out)
	if err != nil {
		return "", err
	}

	return out.Path, nil
}
