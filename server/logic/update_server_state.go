package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"GolandProjects/pbft-gautamsardana/server/storage/datastore"
	"context"
)

func UpdateServerState(ctx context.Context, conf *config.Config, req *common.UpdateServerStateRequest) error {
	conf.MutexLock.Lock()
	defer conf.MutexLock.Unlock()

	datastore.RunDBScript(conf.ServerNumber)

	conf.IsAlive = req.IsAlive
	conf.IsByzantine = req.IsByzantine

	config.InitiateBalance(conf)
	//conf.ViewNumber = 1
	//conf.SequenceNumber = 1
	//conf.NextSequenceNumber = 1
	for k := range conf.PendingTransactions {
		delete(conf.PendingTransactions, k)
	}
	for len(conf.ExecuteSignal) > 0 {
		<-conf.ExecuteSignal
	}
	for k := range conf.PBFTLogs {
		delete(conf.PBFTLogs, k)
	}
	conf.CheckpointRequests = conf.CheckpointRequests[:0]
	return nil
}
