package logic

import (
	common "GolandProjects/pbft/api_common"
	"GolandProjects/pbft/server/config"
	"context"
	"encoding/json"
	"fmt"
)

func Callback(ctx context.Context, conf *config.Config, txn *common.TxnRequest) {
	fmt.Printf("Server %d: sending callback to client\n", conf.ServerNumber)

	txnSignedResp := &common.SignedTxnResponse{
		ViewNumber: conf.ViewNumber,
		TxnID:      txn.TxnID,
		Client:     txn.Sender,
		Status:     StateToStringMap[txn.Status],
	}
	txnSignedRespBytes, err := json.Marshal(txnSignedResp)
	if err != nil {
		fmt.Println(err)
		return
	}
	sign, err := SignMessage(conf.PrivateKey, txnSignedRespBytes)
	if err != nil {
		fmt.Println(err)
		return
	}
	finalResp := &common.TxnResponse{
		SignedMessage: txnSignedRespBytes,
		Sign:          sign,
		ServerNo:      conf.ServerNumber,
	}

	client, clientErr := conf.Pool.GetServer(GetClientAddress())
	if clientErr != nil {
		fmt.Println(clientErr)
		return
	}
	_, err = client.Callback(ctx, finalResp)
	if err != nil {
		fmt.Println(err)
		return
	}
}
