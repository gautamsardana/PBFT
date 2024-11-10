package config

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	publicKeyPool "GolandProjects/pbft-gautamsardana/public_key_pool"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"os"
	"sync"
	"time"

	serverPool "GolandProjects/pbft-gautamsardana/server_pool"
)

const configPath = "/go/src/GolandProjects/pbft-gautamsardana/client/config/config.json"

type Config struct {
	Port             string   `json:"port"`
	ServerCount      int32    `json:"server_count"`
	ServerFaulty     int32    `json:"server_faulty"`
	ServerAddresses  []string `json:"server_addresses"`
	ViewNumber       int32    `json:"view_number"`
	ClientIDs        []string `json:"client_ids"`
	Pool             *serverPool.ServerPool
	PublicKeys       *publicKeyPool.PublicKeyPool
	PrivateKeyString string `json:"private_key"`
	PrivateKey       *rsa.PrivateKey
	TxnQueue         []*common.TxnRequest
	TxnTimeout       time.Duration
	CurrTxn          *common.TxnRequest
	ClientQueues     map[string]*ClientQueue
	QueueMutex       sync.Mutex
	LockMutex        sync.Mutex
}

type ClientQueue struct {
	Queue        []*Transaction
	Mutex        sync.Mutex
	ActiveTxn    *Transaction
	ResponseChan chan *common.TxnResponse
}

type Transaction struct {
	TxnID        string
	Sender       string
	Receiver     string
	Amount       float32
	Timestamp    *timestamppb.Timestamp
	SuccessCount int
	FailureCount int
	ResponseChan chan *common.SignedTxnResponse
	Timer        *time.Timer
	RetryCount   int32
}

func GetConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	jsonConfig, err := os.ReadFile(homeDir + configPath)
	if err != nil {
		log.Fatal(err)
	}
	conf := &Config{}
	if err = json.Unmarshal(jsonConfig, conf); err != nil {
		log.Fatal(err)
	}
	return conf
}

func InitiateConfig(conf *Config) {
	InitiateServerPool(conf)
	InitiatePublicKeys(conf)
	InitiatePrivateKey(conf)
	initiateClientQueues(conf)
}

func initiateClientQueues(conf *Config) {
	if conf.ClientQueues == nil {
		conf.ClientQueues = make(map[string]*ClientQueue)
	}
	conf.QueueMutex.Lock()
	defer conf.QueueMutex.Unlock()
	for _, clientID := range conf.ClientIDs {
		clientQueue := &ClientQueue{
			Queue:        []*Transaction{},
			ResponseChan: make(chan *common.TxnResponse),
		}
		conf.ClientQueues[clientID] = clientQueue
	}
}

func InitiateServerPool(conf *Config) {
	pool, err := serverPool.NewServerPool(conf.ServerAddresses)
	if err != nil {
		//todo: change this
		fmt.Println(err)
	}
	conf.Pool = pool
}

func InitiatePublicKeys(conf *Config) {
	pool, err := publicKeyPool.NewPublicKeyPool()
	if err != nil {
		fmt.Println(err)
	}
	conf.PublicKeys = pool
}

func InitiatePrivateKey(conf *Config) {
	pemData, err := base64.StdEncoding.DecodeString(conf.PrivateKeyString)
	if err != nil {
		return
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return
	}

	conf.PrivateKey = privateKey
}

//func InitiateDKG(conf *Config) {
//	s := &api.Client{Config: conf}
//	_, err := s.InitiateDKG(context.Background(), nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//}
