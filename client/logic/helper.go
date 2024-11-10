package logic

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

var mapLiveServerToServerAddr = map[string]string{
	"S1": "localhost:8080",
	"S2": "localhost:8081",
	"S3": "localhost:8082",
	"S4": "localhost:8083",
	"S5": "localhost:8084",
	"S6": "localhost:8085",
	"S7": "localhost:8086",
}

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
