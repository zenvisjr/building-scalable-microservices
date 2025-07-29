package auth

import (
	"context"
	"fmt"
	"net"

	"github.com/zenvisjr/building-scalable-microservices/account"
	"github.com/zenvisjr/building-scalable-microservices/auth/pb"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedAuthServiceServer
	accountClient *account.Client
	service       Service
}

func ListenGRPC(s Service, port int) error {
	Logs := logger.GetGlobalLogger()
	address := fmt.Sprintf(":%d", port)
	Logs.LocalOnlyInfo("Starting Auth gRPC server on " + address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		Logs.Error(context.Background(), "Failed to bind: "+err.Error())
		return err
	}
	// Connect to Account service
	accountClient, err := account.NewClient("account:8080") // or from env/config
	if err != nil {
		Logs.Error(context.Background(), "Failed to connect to Account Service: "+err.Error())
		return err
	}
	Logs.LocalOnlyInfo("Connected to Account Service")

	server := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryLoggingInterceptor()),
	)

	pb.RegisterAuthServiceServer(server, &grpcServer{
		accountClient: accountClient,
		service:       s,
	})

	reflection.Register(server)
	Logs.LocalOnlyInfo("Auth gRPC server started on " + address)
	Logs.Info(context.Background(), "Auth gRPC server started on "+address)
	return server.Serve(lis)
}

func (g *grpcServer) Signup(ctx context.Context, req *pb.SignupRequest) (*pb.AuthResponse, error) {
	return g.service.Signup(ctx, req.GetName(), req.GetEmail(), req.GetPassword(), req.GetRole(), g.accountClient)
}

func (g *grpcServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	return g.service.Login(ctx, req.GetEmail(), req.GetPassword(), g.accountClient)
}

func (g *grpcServer) RefreshToken(ctx context.Context, req *pb.RefreshRequest) (*pb.AuthResponse, error) {
	return g.service.RefreshToken(ctx, req.GetRefreshToken())
}

func (g *grpcServer) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	Logs := logger.GetGlobalLogger()
	userClaims, err := g.service.VerifyToken(ctx, req.GetAccessToken())
	if err != nil {
		Logs.Error(ctx, "Failed to verify token in server: "+err.Error())
		return nil, err
	}
	return &pb.VerifyResponse{
		UserId: userClaims.ID,
		Email:  userClaims.Email,
		Role:   userClaims.Role,
	}, nil
}
