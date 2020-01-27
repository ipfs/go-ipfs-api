package shell

import (
	"context"
	"encoding/json"

	"github.com/tron-us/go-btfs-common/crypto"
	ledgerpb "github.com/tron-us/go-btfs-common/protos/ledger"
	"github.com/tron-us/go-common/v2/log"

	"go.uber.org/zap"
)

type StorageUploadOpts = func(*RequestBuilder) error

type storageUploadResponse struct {
	SessionId string
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

type ContractItem struct {
	Contract   string
	ContractId int
}

type Contracts struct {
	Contracts []ContractItem `json:contracts`
}

type UnsignedData struct {
	Unsigned string
	Opcode   string
}

func (d UnsignedData) SignData(privateKey string) ([]byte, error) {
	privKey, _ := crypto.ToPrivKey(privateKey)
	signedData, err := privKey.Sign([]byte(d.Unsigned))
	if err != nil {
		return nil, err
	}
	return signedData, nil
}

func (d UnsignedData) SignBalanceData(privateKey string) (*ledgerpb.SignedPublicKey, error) {
	privKey, _ := crypto.ToPrivKey(privateKey)
	pubKeyRaw, err := privKey.GetPublic().Raw()
	if err != nil {
		return &ledgerpb.SignedPublicKey{}, err
	}
	lgPubKey := &ledgerpb.PublicKey{
		Key: pubKeyRaw,
	}
	sig, err := crypto.Sign(privKey, lgPubKey)
	if err != nil {
		return &ledgerpb.SignedPublicKey{}, err
	}
	lgSignedPubKey := &ledgerpb.SignedPublicKey{
		Key:       lgPubKey,
		Signature: sig,
	}
	return lgSignedPubKey, nil
}
func (c Contracts) SignContracts(privateKey string) (*Contracts, error) {
	//do some signing here using private key
	privKey, err := crypto.ToPrivKey(privateKey)
	if err != nil {
		log.Error("%s", zap.Error(err))
	}
	for contractIndex, element := range c.Contracts {
		signedContract, err := privKey.Sign([]byte(element.Contract))
		c.Contracts[contractIndex].Contract = string(signedContract)
		if err != nil {
			return nil, err
		}
	}
	return &c, nil
}

// Set storage upload time.
func StorageLength(length int) StorageUploadOpts {
	return func(rb *RequestBuilder) error {
		rb.Option("storage-length", length)
		return nil
	}
}

// Storage upload api.
func (s *Shell) StorageUpload(hash string, offlinePeerId string, offlineUploadNonceTimestamp string, offlinePeerSessionSignature string, options ...StorageUploadOpts) (string, error) {
	var out storageUploadResponse
	rb := s.Request("storage/upload", hash, offlinePeerId, offlineUploadNonceTimestamp, offlinePeerSessionSignature)
	for _, option := range options {
		_ = option(rb)
	}
	return out.SessionId, rb.Exec(context.Background(), &out)
}

// Storage upload session status api.
func (s *Shell) StorageUploadStatus(id string, offlinePeerId string, offlineUploadNonceTimestamp string, offlinePeerSessionSignature string) (Storage, error) {
	var out Storage
	rb := s.Request("storage/upload/status", id, offlinePeerId, offlineUploadNonceTimestamp, offlinePeerSessionSignature)
	return out, rb.Exec(context.Background(), &out)
}

// Storage upload get offline contract batch api.
func (s *Shell) StorageUploadGetContractBatch(id string, offlinePeerId string, offlineUploadNonceTimestamp string, offlinePeerSessionSignature string, sessionStatus string) (Contracts, error) {
	var out Contracts
	rb := s.Request("storage/upload/getcontractbatch", id, offlinePeerId, offlineUploadNonceTimestamp, offlinePeerSessionSignature, sessionStatus)
	return out, rb.Exec(context.Background(), &out)
}

// Storage upload sign offline contract batch api.
func (s *Shell) StorageUploadSignBatch(id string, offlinePeerId string, offlineUploadNonceTimestamp string, unsignedBatchContracts Contracts, offlinePeerSessionSignature string, sessionStatus string, privateKey string) ([]byte, error) {
	var out []byte

	signedBatchContracts, err := unsignedBatchContracts.SignContracts(privateKey)
	if err != nil {
		log.Error("%s", zap.Error(err))
	}
	byteSignedBatchContracts, err := json.Marshal(signedBatchContracts)
	if err != nil {
		return nil, err
	}
	rb := s.Request("storage/upload/signbatch", id, offlinePeerId, offlineUploadNonceTimestamp, offlinePeerSessionSignature, string(byteSignedBatchContracts))
	return out, rb.Exec(context.Background(), &out)
}

// Storage upload get offline unsigned data api.
func (s *Shell) StorageUploadGetUnsignedData(id string, offlinePeerId string, offlineUploadNonceTimestamp string, offlinePeerSessionSignature string, sessionStatus string) (UnsignedData, error) {
	var out UnsignedData
	rb := s.Request("storage/upload/getunsigned", id, offlinePeerId, offlineUploadNonceTimestamp, offlinePeerSessionSignature, sessionStatus)
	return out, rb.Exec(context.Background(), &out)
}

// Storage upload sign offline data api.
func (s *Shell) StorageUploadSign(id string, offlinePeerId string, offlineUploadNonceTimestamp string, unsignedData UnsignedData, offlinePeerSessionSignature string, sessionStatus string, privateKey string) ([]byte, error) {
	var out []byte
	var rb *RequestBuilder
	signedBytes, err := unsignedData.SignData(privateKey)
	if err != nil {
		log.Error("%s", zap.Error(err))
	}
	rb = s.Request("storage/upload/sign", id, offlinePeerId, offlineUploadNonceTimestamp, offlinePeerSessionSignature, string(signedBytes), sessionStatus)
	return out, rb.Exec(context.Background(), &out)
}

func (s *Shell) StorageUploadSignBalance(id string, offlinePeerId string, offlineUploadNonceTimestamp string, unsignedData UnsignedData, offlinePeerSessionSignature string, sessionStatus string, privateKey string) ([]byte, error) {
	var out []byte
	var rb *RequestBuilder
	ledgerSignedPublicKey, err := unsignedData.SignBalanceData(privateKey)
	if err != nil {
		log.Error("%s", zap.Error(err))
	}
	rb = s.Request("storage/upload/sign", id, offlinePeerId, offlineUploadNonceTimestamp, offlinePeerSessionSignature, ledgerSignedPublicKey.String(), sessionStatus)
	return out, rb.Exec(context.Background(), &out)
}
