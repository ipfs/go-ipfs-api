package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
)

type PublishResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Publish updates a mutable name to point to a given value
func (s *Shell) Publish(node string, value string) error {
	args := []string{value}
	if node != "" {
		args = []string{node, value}
	}

	resp, err := s.newRequest(context.Background(), "name/publish", args...).Send(s.httpcli)
	if err != nil {
		return err
	}
	defer resp.Close()

	if resp.Error != nil {
		return resp.Error
	}

	return nil
}

// PublishWithDetails is used for fine grained control over record publishing
func (s *Shell) PublishWithDetails(contentHash string, lifetime string, ttl string, key string, resolve bool) (*PublishResponse, error) {
	var pubResp PublishResponse
	var resolveString string
	if contentHash == "" {
		return nil, errors.New("empty contentHash provided")
	}
	if lifetime == "" {
		return nil, errors.New("empty lifetime provided")
	}
	if ttl == "" {
		return nil, errors.New("empty ttl provided")
	}
	if key == "" {
		return nil, errors.New("empty key provided")
	}
	if resolve {
		resolveString = "true"
	} else {
		resolveString = "false"
	}
	args := []string{contentHash, resolveString, lifetime, ttl, key}
	resp, err := s.newRequest(context.TODO(), "name/publish", args...).Send(s.httpcli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Output)
	json.Unmarshal(buf.Bytes(), &pubResp)
	return &pubResp, nil
}

// Resolve gets resolves the string provided to an /ipfs/[hash]. If asked to
// resolve an empty string, resolve instead resolves the node's own /ipns value.
func (s *Shell) Resolve(id string) (string, error) {
	var resp *Response
	var err error
	if id != "" {
		resp, err = s.newRequest(context.Background(), "name/resolve", id).Send(s.httpcli)
	} else {
		resp, err = s.newRequest(context.Background(), "name/resolve").Send(s.httpcli)
	}
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
