package logic

import (
	common "GolandProjects/pbft/api_common"
	"context"
	"fmt"
	"time"

	"GolandProjects/pbft/client/config"
)

func InitiateDKG(ctx context.Context, conf *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	for _, serverAddr := range conf.ServerAddresses {
		server, err := conf.Pool.GetServer(serverAddr)
		if err != nil {
			return err
		}
		_, err = server.InitiateDKG(ctx, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

//func PrintBalance(ctx context.Context, req *common.GetBalanceRequest, conf *config.Config) (*common.GetBalanceResponse, error) {
//	serverAddr := mapUserToServer[req.User]
//	server, err := conf.Pool.GetServer(serverAddr)
//	if err != nil {
//		fmt.Println(err)
//	}
//	resp, err := server.PrintBalance(ctx, req)
//	if err != nil {
//		return nil, err
//	}
//	return &common.GetBalanceResponse{
//		Balance: resp.Balance,
//	}, nil
//}
//
//func PrintLogs(ctx context.Context, req *common.PrintLogsRequest, conf *config.Config) (*common.PrintLogsResponse, error) {
//	serverAddr := mapUserToServer[req.User]
//	server, err := conf.Pool.GetServer(serverAddr)
//	if err != nil {
//		fmt.Println(err)
//	}
//	resp, err := server.PrintLogs(ctx, req)
//	if err != nil {
//		return nil, err
//	}
//	return resp, nil
//}
//

func PrintDB(ctx context.Context, req *common.PrintDBRequest, conf *config.Config) (*common.PrintDBResponse, error) {
	serverAddr := mapLiveServerToServerAddr[req.Server]
	server, err := conf.Pool.GetServer(serverAddr)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := server.PrintDB(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//
//func Performance(ctx context.Context, req *common.PerformanceRequest, conf *config.Config) (*common.PerformanceResponse, error) {
//	serverAddr := mapUserToServer[req.User]
//	server, err := conf.Pool.GetServer(serverAddr)
//	if err != nil {
//		fmt.Println(err)
//	}
//	resp, err := server.Performance(ctx, req)
//	if err != nil {
//		return nil, err
//	}
//	return resp, nil
//}
