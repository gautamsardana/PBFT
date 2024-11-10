package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"GolandProjects/pbft-gautamsardana/server/storage/datastore"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

func SendNewView(conf *config.Config) error {
	conf.MutexLock.Lock()
	conf.HasSentNewView[conf.ViewNumber] = true
	conf.MutexLock.Unlock()

	viewChangeMsgs := conf.ViewChange[conf.ViewNumber].ViewChangeRequests
	viewChangeMsgBytes, err := json.Marshal(viewChangeMsgs)
	if err != nil {
		return err
	}

	signedReq := &common.NewViewRequest{
		NewViewNumber:        conf.ViewNumber,
		ViewChangeMessages:   viewChangeMsgBytes,
		StableSequenceNumber: conf.LowWatermark,
	}

	signedReqBytes, err := json.Marshal(signedReq)
	if err != nil {
		return err
	}

	sign, err := SignMessage(conf.PrivateKey, signedReqBytes)
	if err != nil {
		return err
	}

	newViewReq := &common.PBFTCommonRequest{
		SignedMessage: signedReqBytes,
		Sign:          sign,
		ServerNo:      conf.ServerNumber,
	}

	fmt.Printf("Server %d: sending new view req for view %d\n", conf.ServerNumber, conf.ViewNumber)

	var wg sync.WaitGroup

	for _, serverAddress := range conf.ServerAddresses {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			server, serverErr := conf.Pool.GetServer(serverAddress)
			if serverErr != nil {
				fmt.Println(serverErr)
			}
			_, err = server.NewView(context.Background(), newViewReq)
			if err != nil {
				fmt.Println(err)
				return
			}
		}(serverAddress)
	}

	wg.Wait()
	conf.MutexLock.Lock()
	conf.LastViewChangeTime = time.Now()
	conf.IsUnderViewChange[conf.ViewNumber] = false
	conf.MutexLock.Unlock()

	ProcessOldViewTxns(conf)

	return nil
}

func ProcessOldViewTxns(conf *config.Config) {
	viewChangeMsgs := conf.ViewChange[conf.ViewNumber].ViewChangeRequests

	txnMap := make(map[string]*common.TxnRequest)
	isTxnPrepared := make(map[string]bool)

	for _, viewChangeMsg := range viewChangeMsgs {
		for txnID, pbftLog := range viewChangeMsg.PBFTLogs {
			fmt.Println(txnID, len(pbftLog.PrepareRequests))
			if pbftLog.PrepareRequests != nil && len(pbftLog.PrepareRequests) >= int(2*conf.ServerFaulty) {
				txnMap[txnID] = pbftLog.TxnReq
				isTxnPrepared[txnID] = true
				continue
			} else {
				if _, exist := txnMap[txnID]; !exist {
					txnMap[txnID] = pbftLog.TxnReq
				}
			}
		}
	}

	for txnID, txn := range txnMap {
		dbTxn, err := datastore.GetTransactionByTxnID(conf.DataStore, txnID)
		if err != nil && err != sql.ErrNoRows {
			continue
		}
		if dbTxn != nil && (dbTxn.Status == StFailed || dbTxn.Status == StExecuted || dbTxn.Status == StNoOp) {
			continue
		}

		fmt.Printf("Server %d: processing old txn with seq_no %d\n", conf.ServerNumber, txn.SequenceNo)

		conf.PBFTLogsMutex.Lock()
		_, exists := conf.PBFTLogs[txn.TxnID]
		if exists {
			delete(conf.PBFTLogs, txn.TxnID)
		}
		timer := time.NewTimer(3 * time.Second)
		conf.PBFTLogs[txn.TxnID] = config.PBFTLogsInfo{
			TxnReq: txn,
			Timer:  timer,
			Done:   make(chan struct{}),
		}

		conf.PBFTLogsMutex.Unlock()
		go ViewChangeWorker(conf, txn)

		if isTxnPrepared[txnID] {
			err = SendPrePrepare(context.Background(), conf, txn)
			if err != nil {
				fmt.Println(err)
				continue
			}
		} else {
			err = SendNoOP(context.Background(), conf, txn)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func ReceiveNewView(ctx context.Context, conf *config.Config, req *common.PBFTCommonRequest) error {
	if !conf.IsAlive {
		return errors.New("server dead")
	}

	signedReq := &common.NewViewRequest{}
	err := json.Unmarshal(req.SignedMessage, signedReq)
	if err != nil {
		return err
	}
	fmt.Printf("Server %d: received new view req %d\n", conf.ServerNumber, signedReq.NewViewNumber)

	err = VerifyNewView(conf, req)
	if err != nil {
		return err
	}

	conf.MutexLock.Lock()
	conf.SequenceNumber = signedReq.StableSequenceNumber
	conf.ViewNumber = signedReq.NewViewNumber
	conf.IsUnderViewChange[conf.ViewNumber] = false
	conf.LastViewChangeTime = time.Now()
	conf.MutexLock.Unlock()

	return nil
}

func VerifyNewView(conf *config.Config, req *common.PBFTCommonRequest) error {
	serverAddr := MapServerNoToServerAddr[req.ServerNo]
	publicKey, err := conf.PublicKeys.GetPublicKey(serverAddr)
	if err != nil {
		return err
	}

	// Verify the signature on the Accept message
	err = VerifySignature(publicKey, req.SignedMessage, req.Sign)
	if err != nil {
		return err
	}

	signedReq := &common.NewViewRequest{}
	err = json.Unmarshal(req.SignedMessage, signedReq)
	if err != nil {
		return err
	}

	if signedReq.NewViewNumber < conf.ViewNumber {
		return fmt.Errorf("NewViewNumber outdated")
	}

	// majority_check
	if len(signedReq.ViewChangeMessages) < int(2*conf.ServerFaulty+1) {
		fmt.Printf("Server %d: not enough view change requests for this view number:%d, len = %d\n", conf.ServerNumber, conf.ViewNumber, len(signedReq.ViewChangeMessages))
		return fmt.Errorf("not enough view change requests for this view number")
	}

	return nil
}
