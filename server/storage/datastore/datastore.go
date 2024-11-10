package datastore

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/storage"
	"database/sql"
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrNoRowsUpdated = errors.New("no rows updated for user")

func GetBalance(db *sql.DB, user string) (float32, error) {
	var balance float32
	query := `SELECT balance FROM user WHERE user = ?`
	err := db.QueryRow(query, user).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no balance found for user: %d", user)
		}
		return 0, err
	}
	return balance, nil
}

func UpdateBalance(db *sql.DB, user storage.User) error {
	query := `UPDATE user SET balance = ? WHERE user = ?`
	res, err := db.Exec(query, user.Balance, user.User)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrNoRowsUpdated
	}
	return nil
}

func GetTransactionByTxnID(db *sql.DB, txnID string) (*common.TxnRequest, error) {
	transaction := &storage.Transaction{}
	query := `SELECT txn_id, sender, receiver, amount, sequence_no, status, timestamp FROM transaction WHERE txn_id = ?`
	err := db.QueryRow(query, txnID).Scan(
		&transaction.TxnID,
		&transaction.Sender,
		&transaction.Receiver,
		&transaction.Amount,
		&transaction.SequenceNo,
		&transaction.Status,
		&transaction.Timestamp,
	)

	if err != nil {
		return nil, err
	}

	return &common.TxnRequest{
		TxnID:      transaction.TxnID,
		Sender:     transaction.Sender,
		Receiver:   transaction.Receiver,
		Amount:     transaction.Amount,
		SequenceNo: transaction.SequenceNo,
		Status:     transaction.Status,
		Timestamp:  timestamppb.New(transaction.Timestamp),
	}, nil
}

func GetTransactionBySequenceNo(db *sql.DB, seqNo int32) (*common.TxnRequest, error) {
	transaction := &storage.Transaction{}
	query := `SELECT txn_id, sender, receiver, amount, sequence_no, status, timestamp FROM transaction WHERE sequence_no = ?`
	err := db.QueryRow(query, seqNo).Scan(
		&transaction.TxnID,
		&transaction.Sender,
		&transaction.Receiver,
		&transaction.Amount,
		&transaction.SequenceNo,
		&transaction.Status,
		&transaction.Timestamp,
	)

	if err != nil {
		return nil, err
	}

	return &common.TxnRequest{
		TxnID:      transaction.TxnID,
		Sender:     transaction.Sender,
		Receiver:   transaction.Receiver,
		Amount:     transaction.Amount,
		SequenceNo: transaction.SequenceNo,
		Status:     transaction.Status,
		Timestamp:  timestamppb.New(transaction.Timestamp),
	}, nil
}

func InsertTransaction(tx *sql.DB, transaction *common.TxnRequest) error {
	query := `INSERT INTO transaction (txn_id, sender, receiver, amount, sequence_no, status, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.Exec(query, transaction.TxnID, transaction.Sender, transaction.Receiver,
		transaction.Amount, transaction.SequenceNo, transaction.Status, transaction.Timestamp.AsTime())
	if err != nil {
		return err
	}
	return nil
}

func InsertTransactionV2(tx *sql.Tx, transaction *common.TxnRequest) error {
	query := `INSERT INTO transaction (txn_id, sender, receiver, amount, sequence_no, status, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.Exec(query, transaction.TxnID, transaction.Sender, transaction.Receiver,
		transaction.Amount, transaction.SequenceNo, transaction.Status, transaction.Timestamp.AsTime())
	if err != nil {
		return err
	}
	return nil
}

func UpdateTransaction(tx *sql.DB, transaction *common.TxnRequest) error {
	query := `UPDATE transaction SET sequence_no = ?, status = ? WHERE txn_id = ?`
	_, err := tx.Exec(query, transaction.SequenceNo, transaction.Status, transaction.TxnID)
	if err != nil {
		return err
	}
	return nil
}

func GetLatestTermNo(db *sql.DB) (int32, error) {
	query := `SELECT term FROM transaction order by term desc limit 1`
	var term int32
	err := db.QueryRow(query).Scan(&term)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return term, nil
}

func GetTransactionsAfterSequence(db *sql.DB, sequenceNo int32) ([]*common.TxnRequest, error) {
	var transactions []*common.TxnRequest

	query := `SELECT txn_id, sender, receiver, amount, sequence_no, timestamp FROM transaction WHERE sequence_no > ? ORDER BY timestamp`
	rows, err := db.Query(query, sequenceNo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var txn common.TxnRequest
		if err = rows.Scan(&txn.TxnID, &txn.Sender, &txn.Receiver, &txn.Amount, &txn.SequenceNo, &txn.Timestamp); err != nil {
			return nil, err
		}
		transactions = append(transactions, &txn)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func GetAllTransactions(db *sql.DB) ([]*common.TxnRequest, error) {
	var transactions []*common.TxnRequest

	query := `SELECT txn_id, sender, receiver, amount, sequence_no FROM transaction ORDER BY timestamp`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var txn common.TxnRequest
		if err = rows.Scan(&txn.TxnID, &txn.Sender, &txn.Receiver, &txn.Amount, &txn.SequenceNo); err != nil {
			return nil, err
		}
		transactions = append(transactions, &txn)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

//func GetLastSequenceNumber(db *sql.DB) (int32, error) {
//	query := `SELECT sequence_no FROM transaction order by sequence_no desc limit 1`
//	var sequenceNo int32
//	err := db.QueryRow(query).Scan(&sequenceNo)
//	if err != nil {
//		if err == sql.ErrNoRows {
//			return 0, nil
//		}
//		return 0, err
//	}
//	return sequenceNo, nil
//}
