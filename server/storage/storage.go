package storage

import "time"

type User struct {
	User    string
	Balance float32
}

type Transaction struct {
	TxnID      string    `json:"TxnID,omitempty"`
	Sender     string    `json:"Sender,omitempty"`
	Receiver   string    `json:"Receiver,omitempty"`
	Amount     float32   `json:"Amount,omitempty"`
	SequenceNo int32     `json:"SequenceNo,omitempty"`
	Status     int32     `json:"Status,omitempty"`
	Timestamp  time.Time `json:"Timestamp,omitempty"`
}
