package main

import (
	common "GolandProjects/pbft/api_common"
	"bufio"
	"encoding/csv"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"os"
	"strconv"
	"strings"
)

const inputFilePath = "lab1_Test.csv"

var sets map[int32]*common.TxnSet
var activeServers map[string]bool
var totalSets int32

func loadCSV(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record

	sets = make(map[int32]*common.TxnSet)

	var currentSetNo int32
	var currentLiveServers []string
	var currentByzantineServers []string

	for {
		setRow, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading CSV:", err)
			return err
		}

		// Skip empty rows
		if len(setRow) == 0 || (len(setRow) == 1 && strings.TrimSpace(setRow[0]) == "") {
			continue
		}

		// Check if it's a new set (first column is not empty)
		if setRow[0] != "" {
			setNumber, err := strconv.Atoi(setRow[0])
			if err != nil {
				fmt.Println("Error parsing set number:", err)
				continue
			}
			currentSetNo = int32(setNumber)
			totalSets = currentSetNo

			// Ensure there are enough columns before accessing
			if len(setRow) >= 4 {
				currentLiveServers = strings.Split(strings.Trim(setRow[2], "[] "), ",")
				for i := range currentLiveServers {
					currentLiveServers[i] = strings.TrimSpace(currentLiveServers[i])
				}

				currentByzantineServers = strings.Split(strings.Trim(setRow[3], "[] "), ",")
				for i := range currentByzantineServers {
					currentByzantineServers[i] = strings.TrimSpace(currentByzantineServers[i])
				}
			}

			if _, exists := sets[currentSetNo]; !exists {
				sets[currentSetNo] = &common.TxnSet{
					SetNo:            currentSetNo,
					Txns:             []*common.TxnRequest{},
					LiveServers:      currentLiveServers,
					ByzantineServers: currentByzantineServers,
				}
			}
		}

		// Process transactions
		// Ensure there is at least one column
		if len(setRow) >= 2 {
			txnStrings := strings.Split(setRow[1], ";")
			for _, txn := range txnStrings {
				txn = strings.Trim(txn, "() ")
				parts := strings.Split(txn, ",")
				if len(parts) != 3 {
					continue
				}
				amount, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 32)
				if err != nil {
					fmt.Println("Error parsing amount:", err)
					continue
				}

				if sets[currentSetNo] == nil {
					sets[currentSetNo] = &common.TxnSet{
						SetNo: currentSetNo,
						Txns:  []*common.TxnRequest{},
					}
				}

				sets[currentSetNo].Txns = append(sets[currentSetNo].Txns, &common.TxnRequest{
					Sender:   strings.TrimSpace(parts[0]),
					Receiver: strings.TrimSpace(parts[1]),
					Amount:   float32(amount),
				})
			}
		}
	}

	return nil
}

func main() {
	client := InitiateClient()

	err := loadCSV(inputFilePath)
	if err != nil {
		fmt.Println("Error loading CSV:", err)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	var i int32
	for i = 1; i <= totalSets; i++ {
		fmt.Printf("Processing Set %d: Txns: %v LiveServers: %v ByzantineServers: %v\n", i, sets[i].Txns, sets[i].LiveServers, sets[i].ByzantineServers)

		scanner.Scan()
		processSet(sets[i], client)

		for {
			fmt.Println("Type 'next' to process the next set, " +
				"'balance' to get balance, " +
				"'db' to print database, " +
				"'log' to print log," +
				" or 'perf' to print performance")
			scanner.Scan()
			input := scanner.Text()
			if input == "next" {
				break
			} else if input == "db" {
				fmt.Println("Which server? (eg. 'S1' without quotes)")
				scanner.Scan()
				server := scanner.Text()
				printDB(client, server)
			} else if input == "log" {
				fmt.Println("Which user? (eg. 'S1' without quotes)")
				scanner.Scan()
				//user := scanner.Text()
				//printLogs(client, user)
			} else if input == "balance" {
				fmt.Println("Which user? (eg. 'A' without quotes)")
				scanner.Scan()
				//user := scanner.Text()
				//printBalance(client, user)
			} else if input == "perf" {
				fmt.Println("Which user? (eg. 'S1' without quotes)")
				scanner.Scan()
				//user := scanner.Text()
				//performance(client, user)
			} else {
				fmt.Println("Unknown command")
			}
		}
	}
	fmt.Println("All sets processed.")
}

func InitiateClient() common.PBFTClient {
	conn, err := grpc.NewClient("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}
	client := common.NewPBFTClient(conn)
	return client
}
