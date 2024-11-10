package config

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"sync"
	"time"

	publicKeyPool "GolandProjects/pbft-gautamsardana/public_key_pool"
	serverPool "GolandProjects/pbft-gautamsardana/server_pool"
)

const configPath = "/go/src/GolandProjects/pbft-gautamsardana/server/config/config.json"

type Config struct {
	Port                     string `json:"port"`
	ServerNumber             int32  `json:"server_number"`
	ServerTotal              int32  `json:"server_total"`
	ServerFaulty             int32  `json:"server_faulty"`
	Balance                  map[string]float32
	DBCreds                  DBCreds `json:"db_creds"`
	DataStore                *sql.DB
	ServerAddresses          []string `json:"server_addresses"`
	Pool                     *serverPool.ServerPool
	PublicKeys               *publicKeyPool.PublicKeyPool
	PrivateKeyString         string `json:"private_key"`
	PrivateKey               *rsa.PrivateKey
	ViewNumber               int32 `json:"view_number"`
	SequenceNumber           int32 `json:"sequence_number"`
	NextSequenceNumber       int32 `json:"next_sequence_number"`
	SequenceMutex            sync.RWMutex
	Checkpoint               int32 `json:"checkpoint"`
	LowWatermark             int32 `json:"low_watermark"`
	HighWatermark            int32 `json:"high_watermark"`
	K                        int32 `json:"k"`
	PBFTLogs                 map[string]PBFTLogsInfo
	PBFTLogsMutex            sync.RWMutex
	CheckpointRequests       []*DigestCount
	MutexLock                sync.RWMutex
	PendingTransactions      map[int32]*common.TxnRequest
	PendingTransactionsMutex sync.Mutex
	ExecuteSignal            chan struct{}
	IsUnderViewChange        bool
	ViewChange               map[int32]ViewChangeStruct
	HasSentViewChange        map[int32]bool
	LastViewChangeTime       time.Time
	ViewChangeTimer          *time.Timer
	IsAlive                  bool
	IsByzantine              bool
}

type PBFTLogsInfo struct {
	TxnReq           *common.TxnRequest
	PrePrepareDigest string
	PrepareRequests  []*common.PBFTCommonRequest
	AcceptRequests   []*common.PBFTCommonRequest
	Timer            *time.Timer   `json:"-"`
	Done             chan struct{} `json:"-"`
	Perf             time.Time     `json:"-"`
}

type ViewChangeStruct struct {
	ViewChangeRequests []*ViewChangeSignedRequest
}

type DBCreds struct {
	DSN      string `json:"dsn"`
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type DigestCount struct {
	Digest string
	Count  int32
}

type ViewChangeSignedRequest struct {
	ViewNumber           int32
	LastStableCheckpoint int32
	CheckpointMessages   []*DigestCount
	PBFTLogs             map[string]PBFTLogsInfo
	State                []byte
	StateDigest          string
}

type ClientBalance struct {
	Client  string  `json:"client"`
	Balance float32 `json:"balance"`
}

func InitiateConfig(conf *Config) {
	InitiateBalance(conf)
	InitiateServerPool(conf)
	InitiatePublicKeys(conf)
	InitiatePrivateKey(conf)
	InitiatePendingTxnsLogs(conf)
	conf.MutexLock.Lock()
	conf.HasSentViewChange = make(map[int32]bool)
	conf.ViewChange = make(map[int32]ViewChangeStruct)
	ResetPBFTLogs(conf)
	conf.MutexLock.Unlock()
}

func InitiateBalance(conf *Config) {
	conf.Balance = map[string]float32{
		"A": 10,
		"B": 10,
		"C": 10,
		"D": 10,
		"E": 10,
		"F": 10,
		"G": 10,
		"H": 10,
		"I": 10,
		"J": 10,
	}
}

func InitiateServerPool(conf *Config) {
	pool, err := serverPool.NewServerPool(conf.ServerAddresses)
	if err != nil {
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

func InitiatePendingTxnsLogs(conf *Config) {
	conf.SequenceMutex.Lock()
	conf.NextSequenceNumber = 1
	conf.SequenceMutex.Unlock()
	conf.PendingTransactionsMutex.Lock()
	conf.PendingTransactions = make(map[int32]*common.TxnRequest)
	conf.ExecuteSignal = make(chan struct{}, 7)
	conf.PendingTransactionsMutex.Unlock()
}

func ResetPBFTLogs(conf *Config) {
	conf.PBFTLogsMutex.Lock()
	conf.PBFTLogs = map[string]PBFTLogsInfo{}
	conf.CheckpointRequests = []*DigestCount{}
	conf.PBFTLogsMutex.Unlock()
}

func ClearPBFTLogsTillSequence(conf *Config, sequenceNo int32) {
	conf.PBFTLogsMutex.Lock()
	for txnID, txnLog := range conf.PBFTLogs {
		if txnLog.TxnReq.SequenceNo < sequenceNo {
			delete(conf.PBFTLogs, txnID)
		}
	}
	conf.PBFTLogsMutex.Unlock()
}

func GetConfig(filePath string) *Config {
	jsonConfig, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	conf := &Config{}
	if err = json.Unmarshal(jsonConfig, conf); err != nil {
		log.Fatal(err)
	}
	return conf
}

func SetupDB(config *Config) {
	db, err := sql.Open("mysql", config.DBCreds.DSN)
	if err != nil {
		log.Fatal(err)
	}
	config.DataStore = db
	//defer db.Close()
	fmt.Println("MySQL Connected!!")
}
