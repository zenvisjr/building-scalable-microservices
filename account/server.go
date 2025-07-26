package account

import (
	"context"
	"fmt"
	"net"

	"github.com/zenvisjr/building-scalable-microservices/account/pb"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)



type grpcServer struct {
	service Service
	pb.UnimplementedAccountServiceServer
}

func ListenGRPC(s Service, port int) error {
	Logs := logger.GetGlobalLogger()
	address := fmt.Sprintf(":%d", port)
	Logs.LocalOnlyInfo("Initializing gRPC listener for account microservice on " + address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		Logs.Error(context.Background(), "Failed to bind to port: "+err.Error())
		return err
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryLoggingInterceptor()),
	)

	pb.RegisterAccountServiceServer(server, &grpcServer{service: s})
	reflection.Register(server)

	Logs.LocalOnlyInfo("gRPC server for account microservice registered and starting...")
	return server.Serve(lis)
}

func (g *grpcServer) PostAccount(ctx context.Context, req *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Received PostAccount gRPC request for name: " + req.GetName())

	acc, err := g.service.PostAccount(ctx, req.GetName())
	if err != nil {
		Logs.Error(ctx, "PostAccount service error: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "New account created with ID: "+acc.ID)

	return &pb.PostAccountResponse{
		Account: &pb.Account{
			Id:   acc.ID,
			Name: acc.Name,
		},
	}, nil
}

func (g *grpcServer) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Received GetAccount gRPC request for ID: " + req.GetId())

	acc, err := g.service.GetAccount(ctx, req.GetId())
	if err != nil {
		Logs.Error(ctx, "GetAccount service error: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Fetched account with ID: "+acc.ID)

	return &pb.GetAccountResponse{
		Account: &pb.Account{
			Id:   acc.ID,
			Name: acc.Name,
		},
	}, nil
}

func (g *grpcServer) GetAccounts(ctx context.Context, req *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo(fmt.Sprintf("Received GetAccounts request: skip=%d take=%d", req.GetSkip(), req.GetTake()))

	accs, err := g.service.GetAccounts(ctx, req.GetSkip(), req.GetTake())
	if err != nil {
		Logs.Error(ctx, "GetAccounts service error: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, fmt.Sprintf("Fetched %d accounts", len(accs)))

	grpcAccs := make([]*pb.Account, len(accs))
	for i, acc := range accs {
		grpcAccs[i] = &pb.Account{
			Id:   acc.ID,
			Name: acc.Name,
		}
	}

	return &pb.GetAccountsResponse{
		Accounts: grpcAccs,
	}, nil
}
