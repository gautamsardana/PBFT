package logic

import (
	"GolandProjects/pbft-gautamsardana/server/config"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	common "GolandProjects/pbft-gautamsardana/api_common"
)

// When a replica produces a checkpoint, it multicasts a message CHECKPOINT <n,d,i>sig to
//the other replicas, where n is the sequence number of the last request whose execution
//is reflected in the state and d is the digest of the state.

func SendCheckpoint(ctx context.Context, conf *config.Config, txnReq *common.TxnRequest) {
	conf.MutexLock.Lock()
	stateBytes, stateDigest, err := GetState(conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Server %d: sending checkpoint, balanceArr:%v\n", conf.ServerNumber, stateDigest)

	signedReq := &common.SignedMessage{
		ViewNumber:     conf.ViewNumber,
		SequenceNumber: txnReq.SequenceNo,
		Digest:         stateDigest,
	}

	signedReqBytes, err := json.Marshal(signedReq)
	if err != nil {
		fmt.Println(err)
		return
	}
	sign, err := SignMessage(conf.PrivateKey, signedReqBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = AddOwnCheckpoint(conf, stateDigest)
	if err != nil {
		return
	}

	checkpointReq := &common.PBFTCommonRequest{
		SignedMessage: signedReqBytes,
		Sign:          sign,
		Request:       stateBytes,
		ServerNo:      conf.ServerNumber,
	}

	conf.MutexLock.Unlock()
	for _, serverAddress := range conf.ServerAddresses {
		go func(addr string) {
			server, serverErr := conf.Pool.GetServer(serverAddress)
			if serverErr != nil {
				fmt.Println(serverErr)
			}
			_, err = server.Checkpoint(context.Background(), checkpointReq)
			if err != nil {
				fmt.Println(err)
				return
			}
		}(serverAddress)
	}
	conf.MutexLock.Lock()
	for index, checkpointRequest := range conf.CheckpointRequests {
		//majority_check
		if checkpointRequest.Count >= 2*conf.ServerFaulty+1 && index == 0 {
			CheckpointStable(conf, checkpointReq, true)
		} else if checkpointRequest.Count >= 2*conf.ServerFaulty+1 && index != 0 {
			CheckpointStable(conf, checkpointReq, false)
		}
	}
	conf.MutexLock.Unlock()
}

func GetState(conf *config.Config) ([]byte, string, error) {
	order := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	var balanceArr []*common.ClientBalance

	for _, client := range order {
		balance, exists := conf.Balance[client]
		if exists {
			balanceArr = append(balanceArr, &common.ClientBalance{Client: client, Balance: balance})
		}
	}

	balanceArrBytes, err := json.Marshal(balanceArr)
	if err != nil {
		fmt.Println(err)
		return nil, "", err
	}

	digest := sha256.Sum256(balanceArrBytes)
	dHex := fmt.Sprintf("%x", digest[:])
	return balanceArrBytes, dHex, nil
}

func ReceiveCheckpoint(ctx context.Context, conf *config.Config, req *common.PBFTCommonRequest) error {
	if !conf.IsAlive {
		return errors.New("server dead")
	}

	err := VerifyCheckpoint(conf, req)
	if err != nil {
		return err
	}

	conf.MutexLock.Lock()
	AddCheckpoint(conf, req)

	fmt.Printf("Server %d: received checkpoint request:%v\n", conf.ServerNumber, conf.CheckpointRequests)

	for index, checkpointRequest := range conf.CheckpointRequests {
		//majority_check
		if checkpointRequest.Count >= 2*conf.ServerFaulty+1 && index == 0 {
			CheckpointStable(conf, req, true)
		} else if checkpointRequest.Count >= 2*conf.ServerFaulty+1 && index != 0 {
			CheckpointStable(conf, req, false)
		}
	}
	conf.MutexLock.Unlock()
	return nil
}

func VerifyCheckpoint(conf *config.Config, req *common.PBFTCommonRequest) error {
	serverAddr := MapServerNoToServerAddr[req.ServerNo]
	publicKey, err := conf.PublicKeys.GetPublicKey(serverAddr)
	if err != nil {
		return err
	}
	err = VerifySignature(publicKey, req.SignedMessage, req.Sign)
	if err != nil {
		return err
	}
	return nil
}

func AddOwnCheckpoint(conf *config.Config, digest string) error {
	for _, checkpointRequest := range conf.CheckpointRequests {
		if checkpointRequest.Digest == digest {
			checkpointRequest.Count++
			return nil
		}
	}
	conf.CheckpointRequests = append(conf.CheckpointRequests, &config.DigestCount{
		Digest: digest,
		Count:  1,
	})

	return nil
}

func AddCheckpoint(conf *config.Config, req *common.PBFTCommonRequest) {
	signedMessage := &common.SignedMessage{}
	err := json.Unmarshal(req.SignedMessage, signedMessage)
	if err != nil {
		fmt.Println(err)
	}

	for _, checkpointRequest := range conf.CheckpointRequests {
		if checkpointRequest.Digest == signedMessage.Digest {
			checkpointRequest.Count++
			return
		}
		fmt.Println("1", checkpointRequest.Digest, checkpointRequest.Count)
	}
	fmt.Println("2", signedMessage.Digest)
	conf.CheckpointRequests = append(conf.CheckpointRequests, &config.DigestCount{
		Digest: signedMessage.Digest,
		Count:  1,
	})

	return
}

func CheckpointStable(conf *config.Config, req *common.PBFTCommonRequest, isMajority bool) {
	signedMessage := &common.SignedMessage{}
	err := json.Unmarshal(req.SignedMessage, signedMessage)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !isMajority {
		UpdateSequenceNumber(conf, signedMessage.SequenceNumber)
		var receivedState []*common.ClientBalance
		err = json.Unmarshal(req.Request, &receivedState)
		if err != nil {
			fmt.Println(err)
		}
		for _, balance := range receivedState {
			conf.Balance[balance.Client] = balance.Balance
		}
	}

	conf.LowWatermark = signedMessage.SequenceNumber
	conf.HighWatermark = conf.LowWatermark + conf.K
	config.ClearPBFTLogsTillSequence(conf, conf.LowWatermark)
	fmt.Println("checkpoint stable, new lw, hw :%d, %d\n", conf.LowWatermark, conf.HighWatermark)
}
