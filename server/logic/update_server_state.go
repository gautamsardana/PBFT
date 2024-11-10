package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"GolandProjects/pbft-gautamsardana/server/storage/datastore"
	"context"
	"fmt"
	"time"
)

func UpdateServerState(ctx context.Context, conf *config.Config, req *common.UpdateServerStateRequest) error {
	start := time.Now()
	datastore.RunDBScript(conf.ServerNumber)
	fmt.Println(time.Since(start))

	conf.MutexLock.Lock()
	defer conf.MutexLock.Unlock()

	start = time.Now()
	conf.IsAlive = req.IsAlive
	conf.IsByzantine = req.IsByzantine

	conf.ViewNumber = 1
	conf.SequenceNumber = 0
	conf.NextSequenceNumber = 1
	conf.LowWatermark = 0
	conf.HighWatermark = 50

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

	for k := range conf.IsUnderViewChange {
		delete(conf.IsUnderViewChange, k)
	}

	for k := range conf.HasSentViewChange {
		delete(conf.HasSentViewChange, k)
	}

	for k := range conf.HasSentNewView {
		delete(conf.HasSentNewView, k)
	}

	conf.CheckpointRequests = conf.CheckpointRequests[:0]

	fmt.Println(time.Since(start))

	return nil
}
