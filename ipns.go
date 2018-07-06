package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"
	"time"
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
func (s *Shell) PublishWithDetails(contentHash, key string, lifetime, ttl time.Duration, resolve bool) (*PublishResponse, error) {

	args := []string{contentHash}
	req := s.newRequest(context.Background(), "name/publish", args...)
	if key == "" {
		key = "self"
	}
	req.Opts["key"] = key
	if lifetime.Seconds() > 0 {
		req.Opts["lifetime"] = lifetime.String()
	}
	if ttl.Seconds() > 0 {
		req.Opts["ttl"] = ttl.String()
	}
	req.Opts["resolve"] = strconv.FormatBool(resolve)
	resp, err := req.Send(s.httpcli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	if resp.Error != nil {
		return nil, resp.Error
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Output)
	var pubResp PublishResponse
	json.Unmarshal(buf.Bytes(), &pubResp)
	return &pubResp, nil
}

// Resolve gets resolves the string provided to an /ipns/[name]. If asked to
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
