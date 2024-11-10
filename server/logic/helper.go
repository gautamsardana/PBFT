package logic

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/config"
	"GolandProjects/pbft-gautamsardana/server/storage/datastore"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

var MapServerNoToServerAddr = map[int32]string{
	1: "localhost:8080",
	2: "localhost:8081",
	3: "localhost:8082",
	4: "localhost:8083",
	5: "localhost:8084",
	6: "localhost:8085",
	7: "localhost:8086",
}

func SignMessage(privateKey *rsa.PrivateKey, message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func VerifySignature(publicKey *rsa.PublicKey, message, signature []byte) error {
	hash := sha256.Sum256(message)
	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %v", err)
	}
	return nil
}

func IncrementSequenceNumber(conf *config.Config) {
	conf.SequenceMutex.Lock()
	conf.SequenceNumber++
	conf.SequenceMutex.Unlock()
}

func GetSequenceNumber(conf *config.Config) int32 {
	conf.SequenceMutex.RLock()
	defer conf.SequenceMutex.RUnlock()
	return conf.SequenceNumber
}

func UpdateSequenceNumber(conf *config.Config, updatedSequenceNumber int32) {
	conf.SequenceMutex.Lock()
	defer conf.SequenceMutex.Unlock()
	conf.SequenceNumber = updatedSequenceNumber
}

//func ValidateTxn(conf *config.Config, txn *common.TxnRequest) error {
//	dbTxn, err := ValidateTxnInDB(conf, txn)
//	if err != nil && err != sql.ErrNoRows {
//		return err
//	}
//	if dbTxn != nil {
//		err = fmt.Errorf("server %d: duplicate txn found in db", conf.ServerNumber)
//		return err
//	}
//
//	logTxn := ValidateTxnInLogs(conf, txn)
//	if logTxn != nil {
//		err = fmt.Errorf("server %d: duplicate txn found in logs", conf.ServerNumber)
//		return err
//	}
//	return nil
//}

func ValidateTxnInDB(conf *config.Config, req *common.TxnRequest) (*common.TxnRequest, error) {
	txn, err := datastore.GetTransactionByTxnID(conf.DataStore, req.TxnID)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

func GetLeaderNumber(conf *config.Config) int32 {
	leaderNo := conf.ViewNumber % conf.ServerTotal
	if leaderNo == 0 {
		leaderNo = conf.ServerTotal
	}
	return leaderNo
}

func GetClientAddress() string {
	return "localhost:8090"
}
