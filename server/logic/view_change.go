package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func ViewChangeWorker(conf *config.Config, req *common.TxnRequest) {
	conf.PBFTLogsMutex.Lock()
	log := conf.PBFTLogs[req.TxnID]
	conf.PBFTLogsMutex.Unlock()

	defer log.Timer.Stop()

	select {
	case <-log.Timer.C:
		// Check if a view change is already active for the current view number
		conf.MutexLock.Lock()
		//if conf.IsUnderViewChange[conf.ViewNumber] {
		//	fmt.Printf("Server %d: Timer expired, but already under view change for view number %d\n", conf.ServerNumber, conf.ViewNumber)
		//	conf.MutexLock.Unlock()
		//	return
		//}
		if time.Since(conf.LastViewChangeTime) < 2*time.Second {
			fmt.Printf("Server %d: Timer expired, but recent view change occurred, skipping\n", conf.ServerNumber)
			conf.MutexLock.Unlock()
			return
		}
		conf.MutexLock.Unlock()

		// Start a new view change if not already started
		go SendViewChange(conf)
		return
	case <-log.Done:
		fmt.Println("Transaction executed within time, stopping worker for txn:", req.TxnID)
		return
	}
}

func SendViewChange(conf *config.Config) {
	conf.MutexLock.Lock()

	if conf.HasSentViewChange[conf.ViewNumber] && time.Since(conf.LastViewChangeTime) < 2*time.Second {
		fmt.Printf("Server %d: View change already sent for view number %d, skipping\n", conf.ServerNumber, conf.ViewNumber)
		conf.MutexLock.Unlock()
		return
	}

	conf.ViewNumber++
	conf.IsUnderViewChange[conf.ViewNumber] = true
	conf.HasSentViewChange[conf.ViewNumber] = true
	conf.LastViewChangeTime = time.Now()

	fmt.Printf("Server %d: Timer expired, sending view change for view number:%d\n", conf.ServerNumber, conf.ViewNumber)

	// Fetch the state to include in the view change message
	stateBytes, stateDigest, err := GetState(conf)
	if err != nil {
		fmt.Println(err)
		conf.MutexLock.Unlock()
		return
	}
	conf.MutexLock.Unlock()

	signedMessage := &config.ViewChangeSignedRequest{
		ViewNumber:           conf.ViewNumber,
		LastStableCheckpoint: conf.LowWatermark,
		CheckpointMessages:   conf.CheckpointRequests,
		PBFTLogs:             conf.PBFTLogs,
		State:                stateBytes,
		StateDigest:          stateDigest,
	}

	signedMsgBytes, err := json.Marshal(signedMessage)
	if err != nil {
		fmt.Println(err)
		return
	}

	sign, err := SignMessage(conf.PrivateKey, signedMsgBytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	viewChangeReq := &common.PBFTCommonRequest{
		SignedMessage: signedMsgBytes,
		Sign:          sign,
		ServerNo:      conf.ServerNumber,
	}

	// Broadcast view change request to other servers
	for _, serverAddress := range conf.ServerAddresses {
		go func(addr string) {
			server, serverErr := conf.Pool.GetServer(serverAddress)
			if serverErr != nil {
				fmt.Println(serverErr)
			}
			_, err = server.ViewChange(context.Background(), viewChangeReq)
			if err != nil {
				fmt.Println(err)
				return
			}
		}(serverAddress)
	}
}

func ReceiveViewChange(ctx context.Context, conf *config.Config, req *common.PBFTCommonRequest) error {
	if !conf.IsAlive {
		return errors.New("server dead")
	}

	signedMessage := &config.ViewChangeSignedRequest{}
	err := json.Unmarshal(req.SignedMessage, signedMessage)
	if err != nil {
		return err
	}
	fmt.Printf("Server %d: received view change req for view number: %d\n", conf.ServerNumber, signedMessage.ViewNumber)

	if !conf.IsUnderViewChange[conf.ViewNumber] && conf.ViewNumber == signedMessage.ViewNumber {
		return errors.New("view change already happened. Discarding this req")
	}

	conf.MutexLock.Lock()
	conf.IsUnderViewChange[conf.ViewNumber] = true
	conf.MutexLock.Unlock()

	err = VerifyViewChange(conf, req)
	if err != nil {
		return err
	}

	conf.MutexLock.Lock()
	viewChangeLogs, exists := conf.ViewChange[signedMessage.ViewNumber]
	if !exists {
		var viewChangeRequests []*config.ViewChangeSignedRequest
		viewChangeRequests = append(viewChangeRequests, signedMessage)
		conf.ViewChange[signedMessage.ViewNumber] = config.ViewChangeStruct{
			ViewChangeRequests: viewChangeRequests,
		}
	} else {
		viewChangeLogs.ViewChangeRequests = append(viewChangeLogs.ViewChangeRequests, signedMessage)
		conf.ViewChange[signedMessage.ViewNumber] = viewChangeLogs
	}
	conf.MutexLock.Unlock()
	//majority_check
	if len(viewChangeLogs.ViewChangeRequests) >= int(conf.ServerFaulty+1) &&
		!conf.HasSentViewChange[conf.ViewNumber] {
		SendViewChange(conf)
	}

	conf.MutexLock.Lock()
	UpdateStateIfSlow(conf, signedMessage)
	conf.MutexLock.Unlock()

	if GetLeaderNumber(conf) != conf.ServerNumber {
		fmt.Println("not leader or is byzantine")
		return nil
	}

	//majority_check
	fmt.Println("reached", len(viewChangeLogs.ViewChangeRequests))
	fmt.Println(conf.HasSentNewView)
	if len(conf.ViewChange[conf.ViewNumber].ViewChangeRequests) >= int(2*conf.ServerFaulty) {
		if !conf.HasSentNewView[conf.ViewNumber] {
			err = SendNewView(conf)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func VerifyViewChange(conf *config.Config, req *common.PBFTCommonRequest) error {
	serverAddr := MapServerNoToServerAddr[req.ServerNo]
	publicKey, err := conf.PublicKeys.GetPublicKey(serverAddr)
	if err != nil {
		return err
	}
	err = VerifySignature(publicKey, req.SignedMessage, req.Sign)
	if err != nil {
		return err
	}

	signedMessage := &config.ViewChangeSignedRequest{}
	err = json.Unmarshal(req.SignedMessage, signedMessage)
	if err != nil {
		return err
	}
	if signedMessage.ViewNumber < conf.ViewNumber {
		return errors.New("view number outdated")
	}
	//} else if signedMessage.ViewNumber > conf.ViewNumber {
	//	conf.MutexLock.Lock()
	//	conf.ViewNumber = signedMessage.ViewNumber
	//	conf.MutexLock.Unlock()
	//}
	return nil
}

func UpdateStateIfSlow(conf *config.Config, signedMessage *config.ViewChangeSignedRequest) {
	//todo: assumption that the lastCheckpoint from other server is honest
	_, stateDigest, err := GetState(conf)
	if err != nil {
		fmt.Println(err)
		return
	}

	if stateDigest != signedMessage.StateDigest && conf.LowWatermark < signedMessage.LastStableCheckpoint {
		var receivedState []*common.ClientBalance
		err = json.Unmarshal(signedMessage.State, &receivedState)
		if err != nil {
			fmt.Println(err)
		}
		for _, balance := range receivedState {
			conf.Balance[balance.Client] = balance.Balance
		}
		conf.LowWatermark = signedMessage.LastStableCheckpoint
		conf.HighWatermark = conf.LowWatermark + conf.K
	}
}
