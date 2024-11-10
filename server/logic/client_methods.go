package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"GolandProjects/pbft-gautamsardana/server/storage/datastore"
	"context"
	"fmt"
)

func PrintStatus(ctx context.Context, conf *config.Config, req *common.PrintStatusRequest) (*common.PrintStatusResponse, error) {
	dbTxn, err := datastore.GetTransactionBySequenceNo(conf.DataStore, req.SequenceNumber)
	if err != nil {
		return nil, err
	}

	resp := &common.PrintStatusResponse{
		ServerNo: conf.ServerNumber,
		Status:   StateToStringMap[dbTxn.Status],
	}
	return resp, nil
}

func PrintDB(ctx context.Context, conf *config.Config) (*common.PrintDBResponse, error) {
	clients := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	var clientBalance []*common.ClientBalance

	for _, client := range clients {
		balance, err := datastore.GetBalance(conf.DataStore, client)
		if err != nil {
			fmt.Printf("error trying to fetch balance from datastore, err: %v", err)
			return nil, err
		}
		clientBalance = append(clientBalance, &common.ClientBalance{Client: client, Balance: balance})
	}
	return &common.PrintDBResponse{ClientBalance: clientBalance}, nil
}
