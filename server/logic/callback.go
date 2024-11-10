package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	conn, err := grpc.NewClient(GetClientAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
		return
	}
	client := common.NewPBFTClient(conn)

	_, err = client.Callback(ctx, finalResp)
	if err != nil {
		fmt.Println(err)
		return
	}
}
