package api

import (
	"GolandProjects/pbft-gautamsardana/server/config"
	"context"
	"fmt"

	common "GolandProjects/pbft-gautamsardana/api_common"
	"GolandProjects/pbft-gautamsardana/server/logic"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	common.UnimplementedPBFTServer
	Config *config.Config
}

func (s *Server) UpdateServerState(ctx context.Context, req *common.UpdateServerStateRequest) (*emptypb.Empty, error) {
	fmt.Printf("Server %d: IsAlive, IsByzantine set to %t, %t\n", s.Config.ServerNumber, req.IsAlive, req.IsByzantine)

	logic.UpdateServerState(ctx, s.Config, req)

	return nil, nil
}

func (s *Server) ProcessTxn(ctx context.Context, req *common.TxnRequest) (*emptypb.Empty, error) {
	err := logic.ProcessTxn(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("ProcessTxn err: %v\n", err)
		return nil, err
	}
	return nil, nil
}

func (s *Server) PrePrepare(ctx context.Context, req *common.PrePrepareRequest) (*emptypb.Empty, error) {
	err := logic.ReceivePrePrepare(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("PrePrepare err: %v\n", err)
		return nil, err
	}
	return nil, nil
}

func (s *Server) Prepare(ctx context.Context, req *common.PBFTCommonRequest) (*emptypb.Empty, error) {
	err := logic.ReceivePrepare(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("Prepare err: %v\n", err)
		return nil, err
	}
	return nil, nil
}

func (s *Server) Propose(ctx context.Context, req *common.ProposeRequest) (*emptypb.Empty, error) {
	err := logic.ReceivePropose(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("Propose err: %v\n", err)
		return nil, err
	}
	return nil, nil
}

func (s *Server) Accept(ctx context.Context, req *common.PBFTCommonRequest) (*emptypb.Empty, error) {
	err := logic.ReceiveAccept(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("Accept err: %v\n", err)
		return nil, err
	}
	return nil, nil
}

func (s *Server) Commit(ctx context.Context, req *common.CommitRequest) (*emptypb.Empty, error) {
	err := logic.ReceiveCommit(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("Commit err: %v\n", err)
		return nil, err
	}
	return nil, nil
}

func (s *Server) Checkpoint(ctx context.Context, req *common.PBFTCommonRequest) (*emptypb.Empty, error) {
	err := logic.ReceiveCheckpoint(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("Checkpoint err: %v\n", err)
		return nil, err
	}
	return nil, nil
}

func (s *Server) ViewChange(ctx context.Context, req *common.PBFTCommonRequest) (*emptypb.Empty, error) {
	err := logic.ReceiveViewChange(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("ViewChange err: %v\n", err)
		return nil, err
	}
	return nil, nil
}

func (s *Server) NewView(ctx context.Context, req *common.PBFTCommonRequest) (*emptypb.Empty, error) {
	err := logic.ReceiveNewView(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("NewView err: %v\n", err)
		return nil, err
	}
	return nil, nil
}

func (s *Server) NoOp(ctx context.Context, req *common.PBFTCommonRequest) (*emptypb.Empty, error) {
	err := logic.ReceiveNoOp(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("NoOp err: %v\n", err)
		return nil, err
	}
	return nil, err
}

func (s *Server) PrintStatusServer(ctx context.Context, req *common.PrintStatusRequest) (*common.PrintStatusServerResponse, error) {
	resp, err := logic.PrintStatus(ctx, s.Config, req)
	if err != nil {
		fmt.Printf("PrintStatus err: %v\n", err)
		return nil, err
	}
	return resp, nil
}

func (s *Server) PrintDB(ctx context.Context, _ *common.PrintDBRequest) (*common.PrintDBResponse, error) {
	resp, err := logic.PrintDB(ctx, s.Config)
	if err != nil {
		fmt.Printf("PrintDB err: %v\n", err)
		return nil, err
	}
	return resp, nil
}

//func (s *Server) Sync(ctx context.Context, req *common.PBFTCommonRequest) (*common.SyncResponse, error) {
//	resp, err := logic.ReceiveSync(ctx, s.Config, req)
//	if err != nil {
//		fmt.Println(err)
//		return nil, err
//	}
//	return resp, nil
//}
