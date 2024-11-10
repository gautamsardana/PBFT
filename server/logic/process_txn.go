package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"GolandProjects/pbft-gautamsardana/server/storage/datastore"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func ProcessTxn(ctx context.Context, conf *config.Config, req *common.TxnRequest) error {
	fmt.Printf("Server %d: received processtxn request from client:%v\n", conf.ServerNumber, req)
	if !conf.IsAlive {
		return errors.New("server dead")
	}

	if conf.IsUnderViewChange {
		return errors.New("server is under view change")
	}

	err := VerifyRequest(conf, req)
	if err != nil {
		return err
	}

	dbTxn, err := datastore.GetTransactionByTxnID(conf.DataStore, req.TxnID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if dbTxn != nil {
		if dbTxn.Status == StExecuted || dbTxn.Status == StFailed || dbTxn.Status == StNoOp {
			go Callback(context.Background(), conf, dbTxn)
			return errors.New("duplicate txn. Sent status to client")
		}
	}

	if GetLeaderNumber(conf) != conf.ServerNumber {
		go RedirectToLeader(context.Background(), conf, req)
		return errors.New("redirected request to leader")
	}

	timer := time.NewTimer(3 * time.Second)
	conf.PBFTLogsMutex.Lock()
	conf.PBFTLogs[req.TxnID] = config.PBFTLogsInfo{Timer: timer, TxnReq: req, Done: make(chan struct{})}
	conf.PBFTLogsMutex.Unlock()
	go ViewChangeWorker(conf, req)

	IncrementSequenceNumber(conf)
	req.SequenceNo = GetSequenceNumber(conf)

	if dbTxn == nil {
		req.Status = StInit
		err = datastore.InsertTransaction(conf.DataStore, req)
		if err != nil {
			return err
		}
	}

	err = SendPrePrepare(ctx, conf, req)
	if err != nil {
		return err
	}
	return nil
}

func RedirectToLeader(ctx context.Context, conf *config.Config, req *common.TxnRequest) {
	fmt.Printf("Server %d: redirecting to leader for view: %d\n", conf.ServerNumber, conf.ViewNumber)

	serverAddr := MapServerNoToServerAddr[GetLeaderNumber(conf)]
	server, err := conf.Pool.GetServer(serverAddr)
	if err != nil {
		fmt.Println(err)
	}
	_, err = server.ProcessTxn(ctx, req)
	if err != nil {
		fmt.Println(err)
	}
}

func VerifyRequest(conf *config.Config, req *common.TxnRequest) error {
	//publicKey, err := conf.PublicKeys.GetPublicKey(GetClientAddress())
	//if err != nil {
	//	return err
	//}
	//err = VerifySignature(publicKey, req.SignedMessage, req.Sign)
	//if err != nil {
	//	return err
	//}
	//
	//signedMessage := &common.SignedMessage{}
	//err = json.Unmarshal(req.SignedMessage, signedMessage)
	//if err != nil {
	//	return err
	//}

	return nil
}
