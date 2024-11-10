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

// follower received a pre-prepare. Now it sends this prepare back to the leader

func SendPrepare(ctx context.Context, conf *config.Config, req *common.PrePrepareRequest) error {
	fmt.Printf("Server %d: sending prepare to leader\n", conf.ServerNumber)

	sign, err := SignMessage(conf.PrivateKey, req.SignedMessage)
	if err != nil {
		return err
	}

	prepareReq := &common.PBFTCommonRequest{
		SignedMessage: req.SignedMessage,
		Sign:          sign,
		ServerNo:      conf.ServerNumber,
		Request:       req.Request,
	}

	leaderAddr := MapServerNoToServerAddr[req.ServerNo]
	server, err := conf.Pool.GetServer(leaderAddr)
	if err != nil {
		return err
	}
	_, err = server.Prepare(context.Background(), prepareReq)
	if err != nil {
		return err
	}
	return nil
}

// leader receives this prepare from followers

func ReceivePrepare(ctx context.Context, conf *config.Config, req *common.PBFTCommonRequest) error {
	// todo: add timer and majority check
	fmt.Printf("Server %d: received prepare from followers\n", conf.ServerNumber)

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

	err = VerifyPrepare(ctx, conf, req, txnReq)
	if err != nil {
		return err
	}

	conf.PBFTLogsMutex.Lock()
	entry, exists := conf.PBFTLogs[txnReq.TxnID]
	if !exists {
		entry = config.PBFTLogsInfo{}
	}
	entry.PrepareRequests = append(entry.PrepareRequests, req)
	conf.PBFTLogs[txnReq.TxnID] = entry
	conf.PBFTLogsMutex.Unlock()

	dbTxn, err := datastore.GetTransactionByTxnID(conf.DataStore, txnReq.TxnID)
	if err != nil {
		return err
	}
	if dbTxn.Status != StPrePrepared {
		err = fmt.Errorf("Server %d: txn in invalid status, cannot accept prepare, status:%v\n", conf.ServerNumber, dbTxn.Status)
		return err
	}

	return nil
}

func VerifyPrepare(ctx context.Context, conf *config.Config, req *common.PBFTCommonRequest, txnReq *common.TxnRequest) error {
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

	conf.PBFTLogsMutex.RLock()
	//if signedMessage.Digest != conf.PBFTLogs[txnReq.TxnID].PrePrepareDigest {
	//	return errors.New("invalid digest")
	//}
	conf.PBFTLogsMutex.RUnlock()
	if signedMessage.SequenceNumber != txnReq.SequenceNo {
		return errors.New("invalid sequence number")
	}
	if signedMessage.SequenceNumber <= conf.LowWatermark || signedMessage.SequenceNumber > conf.HighWatermark {
		return errors.New("invalid sequence number")
	}

	return nil
}
