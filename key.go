package shell

import "context"

type Key struct {
	Id   string
	Name string
}

type keyListOutput struct {
	Keys []*Key
}

// KeyList List all local keypairs
func (s *Shell) KeyList(ctx context.Context) ([]*Key, error) {
	var out keyListOutput
	if err := s.Request("key/list").Exec(ctx, &out); err != nil {
		return nil, err
	}
	return out.Keys, nil
}
