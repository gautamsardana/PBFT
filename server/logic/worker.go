package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"GolandProjects/pbft-gautamsardana/server/storage"
	"GolandProjects/pbft-gautamsardana/server/storage/datastore"
	"context"
	"fmt"
)

func WorkerProcess(conf *config.Config) {
	for {
		fmt.Println(conf.ExecuteSignal)
		<-conf.ExecuteSignal
		processReadyTransactions(conf)
	}
}

func processReadyTransactions(conf *config.Config) {
	for {
		conf.SequenceMutex.Lock()
		currentSeqNum := conf.NextSequenceNumber
		conf.SequenceMutex.Unlock()

		conf.PendingTransactionsMutex.Lock()
		txnRequest, exists := conf.PendingTransactions[currentSeqNum]
		conf.PendingTransactionsMutex.Unlock()

		if !exists {
			break
		}

		executeTransaction(conf, txnRequest)

		// Update the sequence number
		conf.SequenceMutex.Lock()
		conf.NextSequenceNumber++
		conf.SequenceMutex.Unlock()

		// Remove the executed transaction
		conf.PendingTransactionsMutex.Lock()
		delete(conf.PendingTransactions, currentSeqNum)
		conf.PendingTransactionsMutex.Unlock()
	}
}

func executeTransaction(conf *config.Config, txnRequest *common.TxnRequest) {
	fmt.Printf("executing txn with seq: %s\n", txnRequest.SequenceNo)
	conf.MutexLock.Lock()
	if conf.Balance[txnRequest.Sender] < txnRequest.Amount {
		fmt.Println("insufficient balance")
		txnRequest.Status = StFailed
	} else {
		txnRequest.Status = StExecuted
		conf.Balance[txnRequest.Sender] -= txnRequest.Amount
		conf.Balance[txnRequest.Receiver] += txnRequest.Amount
		fmt.Println("\n", txnRequest, conf.Balance, "\n")
	}

	// Update the transaction in the datastore
	err := datastore.UpdateTransaction(conf.DataStore, txnRequest)
	if err != nil {
		fmt.Printf("Error updating transaction: %v\n", err)
	}

	err = UpdateBalance(conf, txnRequest)
	if err != nil {
		fmt.Printf("Error updating balance: %v\n", err)
	}
	conf.MutexLock.Unlock()

	conf.PBFTLogsMutex.Lock()
	conf.PBFTLogs[txnRequest.TxnID].Done <- struct{}{}
	conf.PBFTLogsMutex.Unlock()

	if txnRequest.SequenceNo%conf.Checkpoint == 0 {
		SendCheckpoint(context.Background(), conf, txnRequest)
	}

	// Callback to client
	go Callback(context.Background(), conf, txnRequest)
}

func UpdateBalance(conf *config.Config, txnRequest *common.TxnRequest) error {
	err := datastore.UpdateBalance(conf.DataStore, storage.User{
		User:    txnRequest.Sender,
		Balance: conf.Balance[txnRequest.Sender],
	})
	if err != nil {
		return err
	}

	err = datastore.UpdateBalance(conf.DataStore, storage.User{
		User:    txnRequest.Receiver,
		Balance: conf.Balance[txnRequest.Receiver],
	})
	if err != nil {
		return err
	}
	return nil
}
