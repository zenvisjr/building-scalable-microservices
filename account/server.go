package account

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.com/zenvisjr/building-scalable-microservices/account/pb"
)

type grpcServer struct {
	service Service
	pb.UnimplementedAccountServiceServer
}

func ListenGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	//register service

	pb.RegisterAccountServiceServer(server, &grpcServer{service: s})

	//register reflection
	reflection.Register(server)

	return server.Serve(lis)

}

func (g *grpcServer) PostAccount(ctx context.Context, req *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	acc, err := g.service.PostAccount(ctx, req.GetName())
	if err != nil {
		return nil, err
	}

	grpcAcc := &pb.Account{
		Id:   acc.ID,
		Name: acc.Name,
	}
	response := &pb.PostAccountResponse{
		Account: grpcAcc,
	}
	return response, nil
}

func (g *grpcServer) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	acc, err := g.service.GetAccount(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	grpcAcc := &pb.Account{
		Id:   acc.ID,
		Name: acc.Name,
	}
	response := &pb.GetAccountResponse{
		Account: grpcAcc,
	}
	return response, nil

}

func (g *grpcServer) GetAccounts(ctx context.Context, req *pb.GetAccountsRequest) (*pb.GetAccountsResponse, error) {
	accs, err := g.service.GetAccounts(ctx, req.GetSkip(), req.GetTake())
	if err != nil {
		return nil, err
	}

	grpcAccs := make([]*pb.Account, len(accs))
	for i, acc := range accs {
		grpcAccs[i] = &pb.Account{
			Id:   acc.ID,
			Name: acc.Name,
		}
	}
	response := &pb.GetAccountsResponse{
		Accounts: grpcAccs,
	}
	return response, nil

}
