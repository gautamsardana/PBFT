package logic

import (
	"GolandProjects/pbft-gautamsardana/client/config"
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"time"

	common "GolandProjects/pbft-gautamsardana/api_common"
)

func ProcessTxnSet(ctx context.Context, conf *config.Config, req *common.TxnSet) error {
	isServerAlive := map[string]bool{
		"localhost:8080": false,
		"localhost:8081": false,
		"localhost:8082": false,
		"localhost:8083": false,
		"localhost:8084": false,
		"localhost:8085": false,
		"localhost:8086": false,
	}

	isServerByzantine := map[string]bool{
		"localhost:8080": false,
		"localhost:8081": false,
		"localhost:8082": false,
		"localhost:8083": false,
		"localhost:8084": false,
		"localhost:8085": false,
		"localhost:8086": false,
	}

	for _, aliveServer := range req.LiveServers {
		isServerAlive[mapLiveServerToServerAddr[aliveServer]] = true
	}

	for _, byzantineServer := range req.ByzantineServers {
		isServerByzantine[mapLiveServerToServerAddr[byzantineServer]] = true
	}

	for _, serverAddr := range conf.ServerAddresses {
		server, err := conf.Pool.GetServer(serverAddr)
		if err != nil {
			fmt.Println(err)
		}
		_, _ = server.UpdateServerState(ctx, &common.UpdateServerStateRequest{
			IsAlive:     isServerAlive[serverAddr],
			IsByzantine: isServerByzantine[serverAddr],
		})
	}

	for _, txn := range req.Txns {
		txnID, err := uuid.NewRandom()
		if err != nil {
			log.Fatalf("failed to generate UUID: %v", err)
		}

		txn.Timestamp = timestamppb.New(time.Now())

		txnReq := &config.Transaction{
			TxnID:        txnID.String(),
			Sender:       txn.Sender,
			Receiver:     txn.Receiver,
			Amount:       txn.Amount,
			Timestamp:    timestamppb.New(time.Now()),
			SuccessCount: 0,
			FailureCount: 0,
			ResponseChan: make(chan *common.SignedTxnResponse, 7),
			Timer:        nil,
		}
		EnqueueTxn(conf, txnReq)
	}
	return nil
}

func EnqueueTxn(conf *config.Config, req *config.Transaction) {
	conf.QueueMutex.Lock()
	clientQueue, exists := conf.ClientQueues[req.Sender]
	conf.QueueMutex.Unlock()

	if !exists {
		fmt.Printf("Client %s does not exist\n", req.Sender)
		return
	}

	clientQueue.Mutex.Lock()
	defer clientQueue.Mutex.Unlock()

	clientQueue.Queue = append(clientQueue.Queue, req)
}
