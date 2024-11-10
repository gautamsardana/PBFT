package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"GolandProjects/pbft-gautamsardana/server/storage/datastore"
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// leader sends propose to followers

func SendPropose(ctx context.Context, conf *config.Config, txnReq *common.TxnRequest) error {
	cert := CreatePreparedCertificate(conf, txnReq)
	pcBytes, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	sign, err := SignMessage(conf.PrivateKey, pcBytes)
	if err != nil {
		return err
	}

	txnReqBytes, err := json.Marshal(txnReq)
	if err != nil {
		return err
	}

	proposeRequest := &common.ProposeRequest{
		PreparedCertificate: pcBytes,
		Sign:                sign,
		ServerNo:            conf.ServerNumber,
		Request:             txnReqBytes,
	}

	fmt.Printf("Server %d: sending propose to followers\n", conf.ServerNumber)

	for _, prepareRequest := range cert.Requests {
		serverAddr := MapServerNoToServerAddr[prepareRequest.ServerNo]
		server, serverErr := conf.Pool.GetServer(serverAddr)
		if serverErr != nil {
			fmt.Println(serverErr)
			continue
		}
		_, err = server.Propose(context.Background(), proposeRequest)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
	return nil
}

func CreatePreparedCertificate(conf *config.Config, req *common.TxnRequest) *common.Certificate {
	conf.PBFTLogsMutex.RLock()
	defer conf.PBFTLogsMutex.RUnlock()

	cert := &common.Certificate{
		ViewNumber:     conf.ViewNumber,
		SequenceNumber: req.SequenceNo,
		Digest:         conf.PBFTLogs[req.TxnID].PrePrepareDigest,
	}

	for _, prepareRequests := range conf.PBFTLogs[req.TxnID].PrepareRequests {
		cert.Requests = append(cert.Requests, prepareRequests)
	}
	return cert
}

// followers receive this from leader and send accept

func ReceivePropose(ctx context.Context, conf *config.Config, req *common.ProposeRequest) error {
	fmt.Printf("Server %d: received propose from leader\n", conf.ServerNumber)

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

	dbTxn, err := datastore.GetTransactionByTxnID(conf.DataStore, txnReq.TxnID)
	if err != nil {
		return err
	}
	if dbTxn.Status != StPrePrepared {
		err = fmt.Errorf("Server %d: txn in invalid status, cannot accept propose, status:%v\n", conf.ServerNumber, dbTxn.Status)
		return err
	}

	err = VerifyPropose(ctx, conf, req, txnReq)
	if err != nil {
		return err
	}

	dbTxn.Status = StPrepared
	err = datastore.UpdateTransaction(conf.DataStore, dbTxn)
	if err != nil {
		return err
	}

	err = SendAccept(ctx, conf, req, dbTxn)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func VerifyPropose(ctx context.Context, conf *config.Config, req *common.ProposeRequest, txnReq *common.TxnRequest) error {
	serverAddr := MapServerNoToServerAddr[req.ServerNo]
	publicKey, err := conf.PublicKeys.GetPublicKey(serverAddr)
	if err != nil {
		return err
	}

	err = VerifySignature(publicKey, req.PreparedCertificate, req.Sign)
	if err != nil {
		return err
	}

	cert := &common.Certificate{}
	err = json.Unmarshal(req.PreparedCertificate, cert)
	if err != nil {
		return err
	}

	conf.MutexLock.Lock()
	pbftLogs := conf.PBFTLogs[txnReq.TxnID]
	pbftLogs.PrepareRequests = cert.Requests
	conf.PBFTLogs[txnReq.TxnID] = pbftLogs
	conf.MutexLock.Unlock()

	validPrepareCount := int32(0)
	for _, prepareRequest := range cert.Requests {
		txnReq = &common.TxnRequest{}
		err = json.Unmarshal(prepareRequest.Request, txnReq)
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = VerifyPrepare(ctx, conf, prepareRequest, txnReq)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if cert.ViewNumber != conf.ViewNumber ||
			cert.SequenceNumber != txnReq.SequenceNo {
			return errors.New("prepared Certificate does not match expected values")
		}
		validPrepareCount++
	}

	//potty_fixed
	if validPrepareCount < 2*conf.ServerFaulty+1 {
		return errors.New("not enough valid prepares")
	}

	return nil
}
