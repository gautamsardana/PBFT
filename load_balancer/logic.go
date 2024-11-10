package main

import (
	common "GolandProjects/pbft-gautamsardana/api_common"
	"context"
	"fmt"
)

//
//func printBalance(client common.PBFTClient, user string) {
//	resp, err := client.PrintDB(context.Background(), &common.GetBalanceRequest{User: user})
//	if err != nil {
//		fmt.Println("Error:", err)
//		return
//	}
//	fmt.Printf("Balance of %s: %.2f\n", user, resp.Balance)
//}

func printDB(client common.PBFTClient, server string) {
	resp, err := client.PrintDB(context.Background(), &common.PrintDBRequest{Server: server})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for _, clientBalance := range resp.ClientBalance {
		fmt.Printf("DB balance for user %s: %.2f\n", clientBalance.Client, clientBalance.Balance)
	}
}

func printStatus(client common.PBFTClient, seqNo int32) {
	resp, err := client.PrintStatusClient(context.Background(), &common.PrintStatusRequest{SequenceNumber: seqNo})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Status for this txn is %v:\n", resp.Status)
}

//
//func printLogs(client common.PBFTClient, user string) {
//	resp, err := client.PrintLogs(context.Background(), &common.PrintLogsRequest{User: user})
//	if err != nil {
//		fmt.Println("Error:", err)
//		return
//	}
//	fmt.Printf("Logs of user %s: %+v\n", user, resp.Logs)
//}
//
//func performance(client common.PBFTClient, user string) {
//	resp, err := client.Performance(context.Background(), &common.PerformanceRequest{User: user})
//	if err != nil {
//		fmt.Println("Error:", err)
//		return
//	}
//	latency := resp.Latency.AsDuration()
//
//	fmt.Printf("Total Latency till now: %s\n", latency)
//	fmt.Printf("Throughput: %.2f transactions/sec\n", resp.Throughput)
//}

func processSet(s *common.TxnSet, client common.PBFTClient) {
	_, err := client.ProcessTxnSet(context.Background(), s)
	if err != nil {
		return
	}
}
