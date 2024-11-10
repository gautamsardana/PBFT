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

	conf.ViewNumber = 1
	conf.SequenceNumber = 0
	conf.NextSequenceNumber = 1
	conf.LowWatermark = 0
	conf.HighWatermark = 50
	conf.IsUnderViewChange = false

	config.InitiateBalance(conf)

	for k := range conf.PendingTransactions {
		delete(conf.PendingTransactions, k)
	}
	for len(conf.ExecuteSignal) > 0 {
		<-conf.ExecuteSignal
	}
	for k := range conf.PBFTLogs {
		delete(conf.PBFTLogs, k)
	}

	for k := range conf.ViewChange {
		delete(conf.ViewChange, k)
	}

	for k := range conf.HasSentViewChange {
		delete(conf.HasSentViewChange, k)
	}

	for k := range conf.HasSentNewView {
		delete(conf.HasSentViewChange, k)
	}

	conf.CheckpointRequests = conf.CheckpointRequests[:0]

	return nil
}
