package shell

import (
	"context"
	"github.com/tron-us/go-btfs-common/crypto"
	"github.com/tron-us/go-common/v2/log"
	"go.uber.org/zap"
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
func (s *Shell) StorageUploadStatus(id string, offlinePeerId int64, offlineUploadNonceTimestamp string, offlinePeerSessionSignature string) (Storage, error) {
	var out Storage
	rb := s.Request("storage/upload/status", id, string(offlinePeerId), offlineUploadNonceTimestamp, offlinePeerSessionSignature)
	return out, rb.Exec(context.Background(), &out)
}

// Storage upload get contract batch api.
func (s *Shell) StorageUploadGetContractBatch(id string, offlinePeerId int64, offlineUploadNonceTimestamp string, offlinePeerSessionSignature string) ([]byte, error) {
	var out []byte
	rb := s.Request("storage/upload/getcontractbatch", id, string(offlinePeerId), offlineUploadNonceTimestamp, offlinePeerSessionSignature)
	return out, rb.Exec(context.Background(), &out)
}
func Sign (privateKey string, unsignedBytes []byte) ([]byte) {
	return ([]byte("Signed contracts"))
}
// Storage upload get contract batch api.
func (s *Shell) StorageUploadSignBatch(id string, offlinePeerId int64, offlineUploadNonceTimestamp string, unsignedBytes []byte, offlinePeerSessionSignature string) ([]byte, error) {
	var out []byte
	privKey, _ := crypto.ToPrivKey("QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv")
	signContrats , err := privKey.Sign(unsignedBytes)
	if err != nil {
		log.Error("%s", zap.Error(err))
	}
	rb := s.Request("storage/upload/signbatch", id, string(offlinePeerId), offlineUploadNonceTimestamp, offlinePeerSessionSignature, string(signContrats))
	return out, rb.Exec(context.Background(), &out)
}
