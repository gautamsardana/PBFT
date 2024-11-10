package main

import (
	"GolandProjects/pbft-gautamsardana/client/logic"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"

	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/client/api"
	"GolandProjects/pbft-gautamsardana/client/config"
)

func main() {
	conf := config.GetConfig()
	config.InitiateConfig(conf)

	for _, clientID := range conf.ClientIDs {
		clientQueue := conf.ClientQueues[clientID]
		go logic.ProcessClientTransactions(conf, clientQueue)
	}
	//logic.TransactionWorker(conf)

	ListenAndServe(conf)
}

func ListenAndServe(conf *config.Config) {
	lis, err := net.Listen("tcp", ":"+conf.Port)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	common.RegisterPBFTServer(s, &api.Client{Config: conf})
	fmt.Printf("gRPC server running on port %v...\n", conf.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
