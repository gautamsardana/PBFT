package main

import (
	"GolandProjects/pbft/server/config"
	"GolandProjects/pbft/server/logic"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"

	common "GolandProjects/pbft/api_common"
	"GolandProjects/pbft/server/api"
)

func main() {
	configPath := flag.String("config", "config1.json", "Path to the configuration file")
	flag.Parse()
	conf := config.GetConfig(*configPath)

	config.SetupDB(conf)
	config.InitiateConfig(conf)
	go logic.WorkerProcess(conf)
	ListenAndServe(conf)
}

func ListenAndServe(conf *config.Config) {
	lis, err := net.Listen("tcp", ":"+conf.Port)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	common.RegisterPBFTServer(s, &api.Server{Config: conf})
	fmt.Printf("gRPC server running on port %v...\n", conf.Port)
	if err = s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
