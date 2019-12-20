package shell

import (
	"context"
)

type storageUploadResponse struct {
	ID string
}

type shard struct {
	Price  int64
	Host   string
	Status string
}

type Storage struct {
	Status   string
	FileHash string
	Shards   map[string]shard
}

// Storage upload api.
func (s *Shell) StorageUpload(hash string) (string, error) {
	var out storageUploadResponse
	rb := s.Request("storage/upload", hash)
	return out.ID, rb.Exec(context.Background(), &out)
}

// Storage upload status api.
func (s *Shell) StorageUploadStatus(id string) (Storage, error) {
	var out Storage
	rb := s.Request("storage/upload/status", id)
	return out, rb.Exec(context.Background(), &out)
}
