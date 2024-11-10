package logic

import (
	common "GolandProjects/pbft/api_common"
	"context"
	"fmt"
	"time"

	"GolandProjects/pbft/client/config"
)

func ProcessClientTransactions(conf *config.Config, clientQueue *config.ClientQueue) {
	for {
		clientQueue.Mutex.Lock()
		if len(clientQueue.Queue) == 0 {
			clientQueue.Mutex.Unlock()
			time.Sleep(500 * time.Millisecond)
			continue
		}

		txn := clientQueue.Queue[0]
		clientQueue.Queue = clientQueue.Queue[1:]
		clientQueue.ActiveTxn = txn
		clientQueue.Mutex.Unlock()

		txn.ResponseChan = make(chan *common.SignedTxnResponse, 7)
		txn.SuccessCount = 0
		txn.FailureCount = 0
		txn.Timer = time.NewTimer(1 * time.Second)

		txnReq := &common.TxnRequest{
			TxnID:      txn.TxnID,
			Sender:     txn.Sender,
			Receiver:   txn.Receiver,
			Amount:     txn.Amount,
			Timestamp:  txn.Timestamp,
			RetryCount: txn.RetryCount,
		}
		fmt.Println(txnReq)
		_ = ProcessTxn(context.Background(), conf, txnReq)

		success := waitForResponses(txn)
		if success {
			fmt.Printf("Transaction %v succeeded\n", txn)
		} else {
			fmt.Printf("Transaction %v failed\n", txn)
		}

		clientQueue.Mutex.Lock()
		clientQueue.ActiveTxn = nil
		clientQueue.Mutex.Unlock()
	}
}

//func TransactionWorker(conf *config.Config) {
//	ticker := time.NewTicker(500 * time.Millisecond)
//
//	go func() {
//		for {
//			select {
//			case <-ticker.C:
//				QueueTransaction(conf)
//			}
//		}
//	}()
//}
//
//func QueueTransaction(conf *config.Config) {
//	conf.QueueMutex.Lock()
//	defer conf.QueueMutex.Unlock()
//
//	if len(conf.TxnQueue) > 0 {
//		conf.RetryCount++
//		txn := conf.TxnQueue[0]
//		conf.CurrTxn = txn
//
//		go func() {
//			err := ProcessTxn(context.Background(), conf, txn)
//			if err != nil {
//				return
//			}
//		}()
//	}
//}
