package logic

import (
	common "GolandProjects/pbft/api_common"
	"GolandProjects/pbft/client/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

const (
	StExecuted = "Executed"
	StFailed   = "Failed"
)

func waitForResponses(txn *config.Transaction) bool {
	defer txn.Timer.Stop()

	for {
		select {
		case response := <-txn.ResponseChan:
			if response.Status == StExecuted {
				txn.SuccessCount++
			} else if response.Status == StFailed {
				txn.FailureCount++
			}

			// potty_fixed
			if txn.SuccessCount >= 5 {
				return true
			}
			if txn.FailureCount >= 5 {
				return false
			}

		case <-txn.Timer.C:
			txn.RetryCount++
			return false
		}
	}
}

func Callback(ctx context.Context, conf *config.Config, resp *common.TxnResponse) error {
	txnResp, err := ValidateTxnResp(ctx, conf, resp)
	if err != nil {
		return err
	}

	fmt.Printf("received callback %v\n", txnResp)

	conf.QueueMutex.Lock()
	clientQueue, exists := conf.ClientQueues[txnResp.Client]
	conf.QueueMutex.Unlock()

	if !exists {
		fmt.Printf("Client %s does not exist\n", txnResp.Client)
		return errors.New("client not found")
	}

	clientQueue.Mutex.Lock()
	activeTxn := clientQueue.ActiveTxn
	clientQueue.Mutex.Unlock()

	if activeTxn == nil || activeTxn.TxnID != txnResp.TxnID {
		fmt.Printf("No active transaction matching ID %s\n", txnResp.TxnID)
		return errors.New("no active transaction")
	}

	conf.LockMutex.Lock()
	conf.ViewNumber = txnResp.ViewNumber
	conf.LockMutex.Unlock()

	select {
	case activeTxn.ResponseChan <- txnResp:
	default:
		fmt.Printf("Response channel for transaction %s is closed or full\n", activeTxn.TxnID)
	}
	return nil
}

func ValidateTxnResp(ctx context.Context, conf *config.Config, req *common.TxnResponse) (*common.SignedTxnResponse, error) {
	serverAddr := MapServerNoToServerAddr[req.ServerNo]
	publicKey, err := conf.PublicKeys.GetPublicKey(serverAddr)
	if err != nil {
		return nil, err
	}
	err = VerifySignature(publicKey, req.SignedMessage, req.Sign)
	if err != nil {
		return nil, err
	}

	signedTxnResponse := &common.SignedTxnResponse{}
	err = json.Unmarshal(req.SignedMessage, signedTxnResponse)
	if err != nil {
		return nil, err
	}
	return signedTxnResponse, nil
}
