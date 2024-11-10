package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/client/config"
	"context"
	"fmt"
)

func ProcessTxn(ctx context.Context, conf *config.Config, req *common.TxnRequest) error {
	if req.RetryCount == 0 {
		SendTxnToLeader(ctx, conf, req)
	} else {
		SendTxnToEveryone(ctx, conf, req)
	}

	return nil
}

func SendTxnToLeader(ctx context.Context, conf *config.Config, req *common.TxnRequest) {
	leaderNo := conf.ViewNumber % conf.ServerCount
	if leaderNo == 0 {
		leaderNo = conf.ServerCount
	}
	leaderAddr := MapServerNoToServerAddr[leaderNo]
	server, serverErr := conf.Pool.GetServer(leaderAddr)
	if serverErr != nil {
		fmt.Println(serverErr)
		return
	}
	_, err := server.ProcessTxn(ctx, req)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func SendTxnToEveryone(ctx context.Context, conf *config.Config, req *common.TxnRequest) {
	for _, serverAddress := range conf.ServerAddresses {
		go func(serverAddress string) {
			server, serverErr := conf.Pool.GetServer(serverAddress)
			if serverErr != nil {
				fmt.Println(serverErr)
				return
			}
			_, err := server.ProcessTxn(ctx, req)
			if err != nil {
				fmt.Println(err)
				return
			}
		}(serverAddress)
	}
}
