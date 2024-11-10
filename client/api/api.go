package api

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"

	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/client/config"
	"GolandProjects/pbft-gautamsardana/client/logic"
)

type Client struct {
	common.UnimplementedPBFTServer
	Config *config.Config
}

func (c *Client) ProcessTxnSet(ctx context.Context, req *common.TxnSet) (*emptypb.Empty, error) {
	err := logic.ProcessTxnSet(ctx, c.Config, req)
	if err != nil {
		log.Printf("Error processing txn from load balancer: %v", err)
		return nil, err
	}
	return nil, nil
}

func (c *Client) Callback(ctx context.Context, resp *common.TxnResponse) (*emptypb.Empty, error) {
	err := logic.Callback(ctx, c.Config, resp)
	if err != nil {
		log.Printf("Error processing callback response: %v", err)
		return nil, err
	}
	return nil, nil
}

//func (c *Client) InitiateDKG(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
//	err := logic.InitiateDKG(ctx, c.Config)
//	if err != nil {
//		log.Printf("Error processing txn from load balancer: %v", err)
//		return nil, err
//	}
//	return nil, nil
//}

//func (c *Client) PrintBalance(ctx context.Context, req *common.GetBalanceRequest) (*common.GetBalanceResponse, error) {
//	resp, err := logic.PrintBalance(ctx, req, c.Config)
//	if err != nil {
//		log.Printf("Error printing balance: %v", err)
//		return nil, err
//	}
//	return resp, nil
//}

//func (c *Client) PrintLogs(ctx context.Context, req *common.PrintLogsRequest) (*common.PrintLogsResponse, error) {
//	resp, err := logic.PrintLogs(ctx, req, c.Config)
//	if err != nil {
//		log.Printf("Error printing logs: %v", err)
//		return nil, err
//	}
//	return resp, nil
//}
//

func (c *Client) PrintDB(ctx context.Context, req *common.PrintDBRequest) (*common.PrintDBResponse, error) {
	resp, err := logic.PrintDB(ctx, req, c.Config)
	if err != nil {
		log.Printf("Error printing DB balances: %v", err)
		return nil, err
	}
	return resp, nil
}

//
//func (c *Client) Performance(ctx context.Context, req *common.PerformanceRequest) (*common.PerformanceResponse, error) {
//	resp, err := logic.Performance(ctx, req, c.Config)
//	if err != nil {
//		log.Printf("Error evaluating performance: %v", err)
//		return nil, err
//	}
//	return resp, nil
//}
