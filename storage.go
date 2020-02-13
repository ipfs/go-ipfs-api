package shell

import (
	"context"
)

type StorageUploadOpts = func(*RequestBuilder) error

type storageUploadResponse struct {
	ID string
}

type shard struct {
	ContractId string
	Price      int64
	Host       string
	Status     string
}

type Storage struct {
	Status   string
	Message  string
	FileHash string
	Shards   map[string]shard
}

// Set storage upload time.
func StorageLength(length int) StorageUploadOpts {
	return func(rb *RequestBuilder) error {
		rb.Option("storage-length", length)
		return nil
	}
}

// Storage upload api.
func (s *Shell) StorageUpload(hash string, options ...StorageUploadOpts) (string, error) {
	var out storageUploadResponse
	rb := s.Request("storage/upload", hash)
	for _, option := range options {
		_ = option(rb)
	}
	return out.ID, rb.Exec(context.Background(), &out)
}

// Storage upload status api.
func (s *Shell) StorageUploadStatus(id string) (Storage, error) {
	var out Storage
	rb := s.Request("storage/upload/status", id)
	return out, rb.Exec(context.Background(), &out)
}
