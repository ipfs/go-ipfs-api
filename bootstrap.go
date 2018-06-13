package shell

import (
	"context"
	"encoding/json"
)

func (s *Shell) BootstrapAdd(peers []string) ([]string, error) {
	resp, err := s.newRequest(context.Background(), "bootstrap/add", peers...).Send(s.httpcli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	addOutput := &struct {
		Peers []string
	}{}

	err = json.NewDecoder(resp.Output).Decode(addOutput)
	if err != nil {
		return nil, err
	}

	return addOutput.Peers, nil
}

func (s *Shell) BootstrapAddDefault() ([]string, error) {
	resp, err := s.newRequest(context.Background(), "bootstrap/add/default").Send(s.httpcli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	if resp.Error != nil {
		return nil, resp.Error
	}

	addOutput := &struct {
		Peers []string
	}{}

	err = json.NewDecoder(resp.Output).Decode(addOutput)
	if err != nil {
		return nil, err
	}

	return addOutput.Peers, nil
}

func (s *Shell) BootstrapRmAll() ([]string, error) {
	resp, err := s.newRequest(context.Background(), "bootstrap/rm/all").Send(s.httpcli)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	rmAllOutput := &struct {
		Peers []string
	}{}

	err = json.NewDecoder(resp.Output).Decode(rmAllOutput)
	if err != nil {
		return nil, err
	}

	return rmAllOutput.Peers, nil
}
