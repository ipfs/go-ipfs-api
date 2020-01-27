package main

import (
	"errors"
	"fmt"
	"time"

	shell "github.com/TRON-US/go-btfs-api"
	"github.com/opentracing/opentracing-go/log"

)
func demoApp(config map[string]string){

	localUrl := "http://localhost:5001"
	s := shell.NewShell(localUrl)
	//upload the hash
	respUpload, err := s.StorageUpload(config["hash"], config["offlinePeerId"], config["offlineSessionSignature"], config["offlineSessionSignature"])
	if err != nil {
		log.Error(err)
	}

	sessionId := respUpload

	for {
		//pool for offline signing status
		uploadResp, statusError := s.StorageUploadStatus(sessionId, config["offlinePeerId"], config["offlineNonceTimestamp"], config["offlineSessionSignature"] )
		if statusError != nil {
			log.Error(statusError)
		}
		switch uploadResp.Status {
		case "InitSignReadyForEscrowStatus", "InitSignReadyForGuardStatus":
			batchContracts, errorUnsignedContracts := s.StorageUploadGetContractBatch(sessionId, config["offlinePeerId"], config["offlineNonceTimestamp"], config["offlineSessionSignature"], uploadResp.Status)
			if errorUnsignedContracts != nil {
				log.Error(errorUnsignedContracts)
			}
			s.StorageUploadSignBatch(sessionId, config["offlinePeerId"], config["offlineNonceTimestamp"], batchContracts, config["offlineSessionSignature"], uploadResp.Status, config["privateKey"])
			time.Sleep(time.Second*10)
			continue
		case "BalanceSignReadyStatus", "PaySignReadyStatus", "GuardSignReadyStatus":
			unsignedData, errorUnsignedContracts := s.StorageUploadGetUnsignedData(sessionId, config["offlinePeerId"], config["offlineNonceTimestamp"], config["offlineSessionSignature"], uploadResp.Status )
			if errorUnsignedContracts != nil {
				log.Error(errorUnsignedContracts)
			}
			switch unsignedData.Opcode{
			case "balance":
				s.StorageUploadSignBalance(sessionId, config["offlinePeerId"], config["offlineNonceTimestamp"], unsignedData, config["offlineSessionSignature"], uploadResp.Status, config["privateKey"])
			case "sign":
				s.StorageUploadSign(sessionId, config["offlinePeerId"], config["offlineNonceTimestamp"], unsignedData , config["offlineSessionSignature"], uploadResp.Status, config["privateKey"])
			}
			time.Sleep(time.Second*10)
			continue
		case "CompleteStatus":
			break
		case "ErrStatus":
			log.Error(errors.New("errStatus: session experienced an error. stopping app"))
			break
		}
		break
	}
}
func main() {
	//need to call the add the file first to obtain the file hash
	configParameters :=  make(map[string]string)
	configParameters["hash"] = "QmWkY8xpySL6GQTSaEh9ZJ2MyWSpzUraUjT18X1Jwvwg2G"
	configParameters["offlinePeerId"] = "16Uiu2HAkwQZvY1mQjWabNr6eDKR7SW5i1KsnVyQpZqWmtRJ6u6SB"
	configParameters["offlineNonceTimestamp"] = time.Now().String()
	configParameters["offlineSessionSignature"] = fmt.Sprintf("%s%s%s", configParameters["offlinePeerId"], configParameters["hash"], configParameters["offlineNonceTimestamp"])
	configParameters["privateKey"] = `CAISINpkyjyl3J7dPQYKkp7YuHrnHRKhfZkf2gkUyhn7Nyej`
	fmt.Println("Starting offline signing demo ... ")
	demoApp(configParameters)
	fmt.Println("Complete offline signing demo ... ")
}
