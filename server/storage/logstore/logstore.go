package logstore

import (
	common "GolandProjects/pbft/api_common"
)

type LogStore struct {
	Logs []*common.TxnRequest
}

func NewLogStore() *LogStore {
	return &LogStore{
		Logs: make([]*common.TxnRequest, 0),
	}
}

func (store *LogStore) AddTransactionLog(txn *common.TxnRequest) {
	store.Logs = append(store.Logs, txn)
}

//func (store *LogStore) GetLatestSequenceNumber() int32 {
//	latestSequenceNumber := int32(0)
//	for _, txn := range store.Logs {
//		if txn.SequenceNo > latestSequenceNumber {
//			latestSequenceNumber = txn.SequenceNo
//		}
//	}
//	return latestSequenceNumber
//}
