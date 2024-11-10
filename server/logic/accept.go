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

// followers receive propose and send accept to leader

func SendAccept(ctx context.Context, conf *config.Config, proposeReq *common.ProposeRequest, txnReq *common.TxnRequest) error {
	conf.PBFTLogsMutex.RLock()
	signedMsg := &common.SignedMessage{
		ViewNumber:     conf.ViewNumber,
		SequenceNumber: txnReq.SequenceNo,
		Digest:         conf.PBFTLogs[txnReq.TxnID].PrePrepareDigest,
	}
	conf.PBFTLogsMutex.RUnlock()
	signedMsgBytes, err := json.Marshal(signedMsg)
	if err != nil {
		return err
	}

	sign, err := SignMessage(conf.PrivateKey, signedMsgBytes)
	if err != nil {
		return err
	}

	acceptMsg := &common.PBFTCommonRequest{
		SignedMessage: signedMsgBytes,
		Sign:          sign,
		Request:       proposeReq.Request,
		ServerNo:      conf.ServerNumber,
	}

	fmt.Printf("Server %d: sending accept to leader\n", conf.ServerNumber)

	leaderAddr := MapServerNoToServerAddr[proposeReq.ServerNo]
	primaryServer, err := conf.Pool.GetServer(leaderAddr)
	if err != nil {
		return err
	}
	_, err = primaryServer.Accept(context.Background(), acceptMsg)
	if err != nil {
		return err
	}

	return nil
}

// leader receives accept from followers and sends commit

func ReceiveAccept(ctx context.Context, conf *config.Config, req *common.PBFTCommonRequest) error {
	fmt.Printf("Server %d: received accept\n", conf.ServerNumber)

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

	conf.PBFTLogsMutex.Lock()
	entry, exists := conf.PBFTLogs[txnReq.TxnID]
	if !exists {
		return errors.New("no pbft logs for this txn")
	}
	entry.AcceptRequests = append(entry.AcceptRequests, req)
	conf.PBFTLogs[txnReq.TxnID] = entry
	conf.PBFTLogsMutex.Unlock()

	dbTxn, err := datastore.GetTransactionByTxnID(conf.DataStore, txnReq.TxnID)
	if err != nil {
		return err
	}

	if dbTxn.Status != StPrepared {
		err = fmt.Errorf("Server %d: txn in invalid status, cannot accept accept, status:%v\n", conf.ServerNumber, dbTxn.Status)
		return err
	}

	err = VerifyAccept(ctx, conf, req, txnReq)
	if err != nil {
		return err
	}

	// todo: if majority reached -
	// will need to send a list of all

	if len(entry.AcceptRequests) >= int(2*conf.ServerFaulty) {
		err = SendCommit(ctx, conf, txnReq)
		if err != nil {
			return err
		}
	}
	return nil
}

func VerifyAccept(ctx context.Context, conf *config.Config, req *common.PBFTCommonRequest, txnReq *common.TxnRequest) error {
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

	signedMessage := &common.SignedMessage{}
	err = json.Unmarshal(req.SignedMessage, signedMessage)
	if err != nil {
		return err
	}

	conf.PBFTLogsMutex.RLock()
	defer conf.PBFTLogsMutex.RUnlock()

	//if signedMessage.ViewNumber != conf.ViewNumber ||
	//	signedMessage.SequenceNumber != txnReq.SequenceNo ||
	//	signedMessage.Digest != conf.PBFTLogs[txnReq.TxnID].PrePrepareDigest {
	//	return errors.New("accept message does not match expected values")
	//}

	return nil
}
