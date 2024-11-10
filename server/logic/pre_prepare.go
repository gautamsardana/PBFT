package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"GolandProjects/pbft-gautamsardana/server/storage/datastore"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func SendPrePrepare(ctx context.Context, conf *config.Config, req *common.TxnRequest) error {
	//config.NewAcceptedServersInfo(conf) --
	//todo - instead clear the prepared and accepted requests from conf
	fmt.Printf("Server %d: sending pre prepare to followers\n", conf.ServerNumber)

	requestBytes, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	digest := sha256.Sum256(requestBytes)
	dHex := fmt.Sprintf("%x", digest[:])

	signedReq := &common.SignedMessage{
		ViewNumber:     conf.ViewNumber,
		SequenceNumber: req.SequenceNo,
		Digest:         dHex,
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

	prePrepareReq := &common.PrePrepareRequest{
		SignedMessage: signedReqBytes,
		Sign:          sign,
		Request:       requestBytes,
		ServerNo:      conf.ServerNumber,
	}

	conf.PBFTLogsMutex.Lock()
	log, exists := conf.PBFTLogs[req.TxnID]
	if exists {
		log.PrePrepareDigest = dHex

	} else {
		log = config.PBFTLogsInfo{PrePrepareDigest: dHex}
	}
	conf.PBFTLogs[req.TxnID] = log
	conf.PBFTLogsMutex.Unlock()

	req.Status = StPrePrepared
	err = datastore.UpdateTransaction(conf.DataStore, req)
	if err != nil {
		return err
	}

	for _, serverAddress := range conf.ServerAddresses {
		server, serverErr := conf.Pool.GetServer(serverAddress)
		if serverErr != nil {
			fmt.Println(serverErr)
		}
		_, err = server.PrePrepare(context.Background(), prePrepareReq)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReceivePrePrepare(ctx context.Context, conf *config.Config, req *common.PrePrepareRequest) error {
	//todo: checks for duplicate requests
	fmt.Printf("Server %d: received pre prepare from leader\n", conf.ServerNumber)
	if !conf.IsAlive {
		return errors.New("server dead")
	}

	if conf.IsUnderViewChange {
		return errors.New("server is under view change")
	}

	txnReq := &common.TxnRequest{}
	err := json.Unmarshal(req.Request, txnReq)
	if err != nil {
		return err
	}

	err = VerifyPrePrepare(ctx, conf, req, txnReq)
	if err != nil {
		return err
	}

	signedMessage := &common.SignedMessage{}
	err = json.Unmarshal(req.SignedMessage, signedMessage)
	if err != nil {
		return err
	}

	timer := time.NewTimer(1 * time.Second)
	conf.PBFTLogsMutex.Lock()
	conf.PBFTLogs[txnReq.TxnID] = config.PBFTLogsInfo{
		TxnReq:           txnReq,
		PrePrepareDigest: signedMessage.Digest,
		Timer:            timer,
		Done:             make(chan struct{}),
	}

	go ViewChangeWorker(conf, txnReq)
	conf.PBFTLogsMutex.Unlock()

	dbTxn, err := datastore.GetTransactionByTxnID(conf.DataStore, txnReq.TxnID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if dbTxn == nil {
		txnReq.Status = StPrePrepared
		err = datastore.InsertTransaction(conf.DataStore, txnReq)
		if err != nil {
			return err
		}
	}

	err = SendPrepare(ctx, conf, req)
	if err != nil {
		return err
	}

	return nil
}

func VerifyPrePrepare(ctx context.Context, conf *config.Config, req *common.PrePrepareRequest, txnReq *common.TxnRequest) error {
	serverAddr := MapServerNoToServerAddr[req.ServerNo]
	publicKey, err := conf.PublicKeys.GetPublicKey(serverAddr)
	if err != nil {
		return err
	}
	err = VerifySignature(publicKey, req.SignedMessage, req.Sign)
	if err != nil {
		return err
	}

	signedMessage := &common.SignedMessage{}
	err = json.Unmarshal(req.SignedMessage, signedMessage)
	if err != nil {
		return err
	}
	if signedMessage.ViewNumber != conf.ViewNumber {
		return errors.New("invalid view number")
	}

	if signedMessage.SequenceNumber <= conf.LowWatermark || signedMessage.SequenceNumber > conf.HighWatermark {
		return errors.New("invalid sequence")
	}

	//todo - might not need this check
	if signedMessage.SequenceNumber != txnReq.SequenceNo {
		return errors.New("invalid sequence number")
	}

	digest := sha256.Sum256(req.Request)
	dHex := fmt.Sprintf("%x", digest[:])
	if dHex != signedMessage.Digest {
		return errors.New("invalid digest")
	}
	UpdateSequenceNumber(conf, signedMessage.SequenceNumber)

	return nil
}
