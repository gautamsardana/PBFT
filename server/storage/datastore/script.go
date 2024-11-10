package datastore

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var MapServerNoToDBName = map[int32]string{
	1: "S1",
	2: "S2",
	3: "S3",
	4: "S4",
	5: "S5",
	6: "S6",
	7: "S7",
}

const (
	host     = "localhost"
	user     = "root"
	password = ""
)

// Database names
var databases = []string{"s1", "s2", "s3", "s4", "s5", "s6", "s7"}

func DeleteTransactions(db *sql.DB) error {
	// Execute delete statement
	_, err := db.Exec("DELETE FROM transaction")
	if err != nil {
		return fmt.Errorf("could not delete entries in 'transaction' table: %v", err)
	}
	log.Println("All entries deleted from 'transaction' table")
	return nil
}

func UpdateBalances(db *sql.DB) error {
	// Update balances for users 'A' through 'J' to 10
	for ch := 'A'; ch <= 'J'; ch++ {
		_, err := db.Exec("UPDATE user SET balance = 10 WHERE user = ?", string(ch))
		if err != nil {
			return fmt.Errorf("could not update balance for user %c: %v", ch, err)
		}
	}
	log.Printf("Balance updated for all users")
	return nil
}

func ProcessDatabase(dbName string) error {
	// Data Source Name (DSN) format: "user:password@tcp(host)/dbname"
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, dbName)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("could not connect to database %s: %v", dbName, err)
	}
	defer db.Close()

	// Delete transactions and update balances
	if err := DeleteTransactions(db); err != nil {
		return err
	}
	if err := UpdateBalances(db); err != nil {
		return err
	}

	log.Printf("All transactions deleted and balances updated in database '%s'\n", dbName)
	return nil
}

func RunDBScript(serverNo int32) {
	if err := ProcessDatabase(MapServerNoToDBName[serverNo]); err != nil {
		log.Printf("Error: %v\n", err)
	}

}
