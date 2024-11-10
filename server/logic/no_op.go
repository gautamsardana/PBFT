package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"GolandProjects/pbft-gautamsardana/server/storage/datastore"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

func SendNoOP(ctx context.Context, conf *config.Config, req *common.TxnRequest) error {
	fmt.Printf("Server %d: sending no op to followers\n", conf.ServerNumber)

	dbTxn, err := datastore.GetTransactionByTxnID(conf.DataStore, req.TxnID)
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		req.Status = StNoOp
		err = datastore.InsertTransaction(conf.DataStore, req)
		if err != nil {
			return err
		}
	} else {
		dbTxn.Status = StNoOp
		err = datastore.UpdateTransaction(conf.DataStore, dbTxn)
		if err != nil {
			return err
		}
	}

	requestBytes, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	signedReq := &common.SignedMessage{
		ViewNumber:     conf.ViewNumber,
		SequenceNumber: req.SequenceNo,
	}

	signedReqBytes, err := json.Marshal(signedReq)
	if err != nil {
		fmt.Println(err)
		return err
	}
	sign, err := SignMessage(conf.PrivateKey, signedReqBytes)
	if err != nil {
		fmt.Println(err)
		return err
	}

	noOpReq := &common.PBFTCommonRequest{
		SignedMessage: signedReqBytes,
		Sign:          sign,
		Request:       requestBytes,
		ServerNo:      conf.ServerNumber,
	}

	for _, serverAddress := range conf.ServerAddresses {
		go func(addr string) {
			server, serverErr := conf.Pool.GetServer(serverAddress)
			if serverErr != nil {
				fmt.Println(serverErr)
			}
			_, err = server.NoOp(context.Background(), noOpReq)
			if err != nil {
				fmt.Println(err)
				return
			}
		}(serverAddress)
	}
	return nil
}

func ReceiveNoOp(ctx context.Context, conf *config.Config, req *common.PBFTCommonRequest) error {
	fmt.Printf("Server %d: received no op from leader\n", conf.ServerNumber)

	if !conf.IsAlive {
		return errors.New("server dead")
	}

	err := VerifyNoOp(conf, req)
	if err != nil {
		return err
	}

	txnReq := &common.TxnRequest{}
	err = json.Unmarshal(req.Request, txnReq)
	if err != nil {
		return err
	}

	dbTxn, err := datastore.GetTransactionByTxnID(conf.DataStore, txnReq.TxnID)
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		txnReq.Status = StNoOp
		err = datastore.InsertTransaction(conf.DataStore, txnReq)
		if err != nil {
			return err
		}
	} else {
		dbTxn.Status = StNoOp
		err = datastore.UpdateTransaction(conf.DataStore, dbTxn)
		if err != nil {
			return err
		}
	}
	return nil
}

func VerifyNoOp(conf *config.Config, req *common.PBFTCommonRequest) error {
	serverAddr := MapServerNoToServerAddr[req.ServerNo]
	publicKey, err := conf.PublicKeys.GetPublicKey(serverAddr)
	if err != nil {
		return err
	}

	// Verify the signature on the Accept message
	err = VerifySignature(publicKey, req.SignedMessage, req.Sign)
	if err != nil {
		return err
	}

	signedReq := &common.SignedMessage{}
	err = json.Unmarshal(req.SignedMessage, signedReq)
	if err != nil {
		return err
	}

	if signedReq.ViewNumber != conf.ViewNumber {
		return fmt.Errorf("invalid view number")
	}
	return nil
}
