package logic

import (
	common "GolandProjects/pbft/api_common"
	"GolandProjects/pbft/server/config"
	"GolandProjects/pbft/server/storage/datastore"
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

func SendCommit(ctx context.Context, conf *config.Config, proposeReq *common.PBFTCommonRequest, txnReq *common.TxnRequest) error {
	// req can come after receiving prepares OR accepts

	cert := CreateCommitCertificate(conf, txnReq)
	certBytes, err := json.Marshal(cert)
	if err != nil {
		return err
	}

	sign, err := SignMessage(conf.PrivateKey, certBytes)
	if err != nil {
		return err
	}

	commitReq := &common.CommitRequest{
		CommitCertificate: certBytes,
		Sign:              sign,
		Request:           proposeReq.Request,
		ServerNo:          conf.ServerNumber,
	}

	fmt.Printf("Server %d: sending commit to followers\n", conf.ServerNumber)

	for _, acceptRequest := range cert.Requests {
		serverAddr := MapServerNoToServerAddr[acceptRequest.ServerNo]
		server, serverErr := conf.Pool.GetServer(serverAddr)
		if serverErr != nil {
			fmt.Println(serverErr)
			continue
		}

		_, err = server.Commit(context.Background(), commitReq)
		if err != nil {
			fmt.Println(err)
		}
	}

	dbTxn, err := datastore.GetTransactionByTxnID(conf.DataStore, txnReq.TxnID)
	if err != nil {
		return err
	}
	dbTxn.Status = StCommitted
	err = datastore.UpdateTransaction(conf.DataStore, dbTxn)
	if err != nil {
		return err
	}

	err = ExecuteTxn(ctx, conf, txnReq)
	if err != nil {
		return err
	}
	// todo main view change timer stop

	return nil
}

func CreateCommitCertificate(conf *config.Config, txnReq *common.TxnRequest) *common.Certificate {
	conf.PBFTLogsMutex.RLock()
	defer conf.PBFTLogsMutex.RUnlock()

	cert := &common.Certificate{
		ViewNumber:     conf.ViewNumber,
		SequenceNumber: txnReq.SequenceNo,
		Digest:         conf.PBFTLogs[txnReq.TxnID].PrePrepareDigest,
	}

	for _, acceptRequests := range conf.PBFTLogs[txnReq.TxnID].AcceptRequests {
		cert.Requests = append(cert.Requests, acceptRequests)
	}
	return cert
}

func ReceiveCommit(ctx context.Context, conf *config.Config, req *common.CommitRequest) error {
	fmt.Printf("Server %d: received commit from leader\n", conf.ServerNumber)

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

	if dbTxn.Status != StPrepared {
		err = fmt.Errorf("Server %d: txn in invalid status, cannot accept commit, status:%v\n", conf.ServerNumber, dbTxn.Status)
		return err
	}

	err = VerifyCommit(ctx, conf, req, txnReq)
	if err != nil {
		return err
	}

	dbTxn.Status = StCommitted
	err = datastore.UpdateTransaction(conf.DataStore, dbTxn)
	if err != nil {
		return err
	}

	err = ExecuteTxn(ctx, conf, txnReq)
	if err != nil {
		return err
	}

	return nil
}

func VerifyCommit(ctx context.Context, conf *config.Config, req *common.CommitRequest, txnReq *common.TxnRequest) error {
	serverAddr := MapServerNoToServerAddr[req.ServerNo]
	publicKey, err := conf.PublicKeys.GetPublicKey(serverAddr)
	if err != nil {
		return err
	}

	err = VerifySignature(publicKey, req.CommitCertificate, req.Sign)
	if err != nil {
		return err
	}

	cert := &common.Certificate{}
	err = json.Unmarshal(req.CommitCertificate, cert)
	if err != nil {
		return err
	}

	validAcceptCount := int32(0)
	for _, acceptRequest := range cert.Requests {
		acceptTxnReq := &common.TxnRequest{}
		err = json.Unmarshal(req.Request, acceptTxnReq)
		if err != nil {
			return err
		}
		err = VerifyAccept(ctx, conf, acceptRequest, acceptTxnReq)
		if err != nil {
			fmt.Println(err)
			continue
		}
		validAcceptCount++
	}

	//potty_fixed
	if validAcceptCount < 2*conf.ServerFaulty+1 {
		return errors.New("not enough valid accepts")
	}

	if cert.ViewNumber != conf.ViewNumber ||
		cert.SequenceNumber != txnReq.SequenceNo ||
		cert.Digest != conf.PBFTLogs[txnReq.TxnID].PrePrepareDigest {
		return errors.New("commit certificate does not match expected values")
	}
	return nil
}

func ExecuteTxn(ctx context.Context, conf *config.Config, txnRequest *common.TxnRequest) error {
	conf.PendingTransactionsMutex.Lock()
	conf.PendingTransactions[txnRequest.SequenceNo] = txnRequest
	conf.PendingTransactionsMutex.Unlock()

	// Signal the worker unconditionally
	select {
	case conf.ExecuteSignal <- struct{}{}:
	default:
		fmt.Println("// Worker already signaled or signal channel is full")
	}
	return nil
}

//func ExecuteTxn(ctx context.Context, conf *config.Config, txnRequest *common.TxnRequest, isSync bool) error {
//	conf.MutexLock.Lock()
//	if conf.Balance[txnRequest.Sender] < txnRequest.Amount {
//		fmt.Println("insufficient balance")
//		txnRequest.Status = StFailed
//	} else {
//		txnRequest.Status = StExecuted
//		conf.Balance[txnRequest.Sender] -= txnRequest.Amount
//		conf.Balance[txnRequest.Receiver] += txnRequest.Amount
//		conf.SequenceMutex.Lock()
//		conf.LastSequenceNumberLog = txnRequest.SequenceNo
//		conf.SequenceMutex.Unlock()
//	}
//	conf.MutexLock.Unlock()
//
//	err := datastore.UpdateTransaction(conf.DataStore, txnRequest)
//	if err != nil {
//		return err
//	}
//	if !isSync {
//		go Callback(context.Background(), conf, txnRequest)
//	}
//
//	return nil
//}
