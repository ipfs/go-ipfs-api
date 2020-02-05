package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/TRON-US/go-btfs-api/utils"
	"time"

	shell "github.com/TRON-US/go-btfs-api"

	"github.com/opentracing/opentracing-go/log"

)

func demoApp(demoState chan string){

	//defer close(demoState)

	//localUrl := "http://localhost:5001"
	localUrl := "https://storageupload.free.beeceptor.com"
	s := shell.NewShell(localUrl)

	rand := utils.RandString(32)

	//upload a random data and retrieve the hash
	mhash, err := s.Add(bytes.NewBufferString(rand), shell.OnlyHash(true))
	if err != nil {
		log.Error(err)
	}

	//upload the hash
	sessionId, err := s.StorageUpload(mhash)
	if err != nil {
		log.Error(err)
	}

	for {
		//pool for offline signing status
		uploadResp, statusError := s.StorageUploadStatus(sessionId, mhash)
		if statusError != nil {
			log.Error(statusError)
		}
		switch uploadResp.Status {
		case "Uninitialized":
			demoState <- uploadResp.Status
			time.Sleep(time.Second*10)
			continue
		case "initSignReadyForEscrow", "initSignReadyForGuard":
			demoState <- uploadResp.Status
			batchContracts, errorUnsignedContracts := s.StorageUploadGetContractBatch(sessionId, mhash, uploadResp.Status)
			if errorUnsignedContracts != nil {
				log.Error(errorUnsignedContracts)
			}
			s.StorageUploadSignBatch(sessionId, mhash, batchContracts, uploadResp.Status)
			time.Sleep(time.Second*10)
			continue
		case "balanceSignReady", "payChannelSignReady", "payRequestSignReady", "guardSignReady":
			demoState <- uploadResp.Status
			unsignedData, errorUnsignedContracts := s.StorageUploadGetUnsignedData(sessionId, mhash, uploadResp.Status )
			if errorUnsignedContracts != nil {
				log.Error(errorUnsignedContracts)
			}
			switch unsignedData.Opcode{
			case "balance":
				s.StorageUploadSignBalance(sessionId, mhash, unsignedData, uploadResp.Status)
			case "paychannel":
				s.StorageUploadSignPayChannel(sessionId, mhash, unsignedData, uploadResp.Status, unsignedData.Price)
			case "payrequest":
				s.StorageUploadSignPayRequest(sessionId, mhash, unsignedData, uploadResp.Status)
			case "sign":
				s.StorageUploadSign(sessionId, mhash, unsignedData, uploadResp.Status)
			}
			time.Sleep(time.Second*10)
			continue
		case "retrySignReady":
			demoState <- uploadResp.Status
			batchContracts, errorUnsignedContracts := s.StorageUploadGetContractBatch(sessionId,mhash, uploadResp.Status)
			if errorUnsignedContracts != nil {
				log.Error(errorUnsignedContracts)
			}
			s.StorageUploadSignBatch(sessionId, mhash, batchContracts, uploadResp.Status)
			time.Sleep(time.Second*10)
			continue
		case "retrySignProcess":
			demoState <- uploadResp.Status
			time.Sleep(time.Second*10)
			continue
		case "init":
			demoState <- uploadResp.Status
			time.Sleep(time.Second*10)
			continue
		case "complete":
			demoState <- uploadResp.Status
			break
		case "error":
			demoState <- uploadResp.Status
			log.Error(errors.New("errStatus: session experienced an error. stopping app"))
			break
		}
		break
	}
}
func main() {
	//need to call the add the file first to obtain the file hash
	demoState := make(chan string)
	fmt.Println("Starting offline signing demo ... ")
	go demoApp(demoState)
	for {
		select {
		case i := <- demoState:
			fmt.Println("Current status of offline signing demo: " + i)
		case <-time.After(30*time.Second): //simulates timeout
			fmt.Println("Time out: No news in 30 seconds")
		}
	}
	fmt.Println("Complete offline signing demo ... ")
}
