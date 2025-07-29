package account

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/nats-io/nats.go"
	"github.com/zenvisjr/building-scalable-microservices/account/pb"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	service Service
	// mailClient *mail.Mail
	pb.UnimplementedAccountServiceServer
	netScan *nats.Conn
}

func ListenGRPC(s Service, mailURL string, port int) error {
	Logs := logger.GetGlobalLogger()
	address := fmt.Sprintf(":%d", port)
	Logs.LocalOnlyInfo("Initializing gRPC listener for account microservice on " + address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		Logs.Error(context.Background(), "Failed to bind to port: "+err.Error())
		return err
	}

	// mailClient, err := mail.NewMailClient(mailURL)
	// if err != nil {
	// 	Logs.Error(context.Background(), "Failed to connect to Mail gRPC: "+err.Error())
	// 	return err
	// }
	// Logs.LocalOnlyInfo("Connected to Mail service: " + mailURL)

	//add nats server for queuing, we will send mail to this insted of sending direct
	Logs.LocalOnlyInfo("Connecting to NATS in account microservice")

	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		Logs.Error(context.Background(), "Failed to connect to NATS: "+err.Error())
		return err
	}
	Logs.LocalOnlyInfo("Connected to NATS")
	server := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryLoggingInterceptor()),
	)

	pb.RegisterAccountServiceServer(server, &grpcServer{
		service: s,
		// mailClient: mailClient,
		netScan: nc,
	})
	reflection.Register(server)

	Logs.LocalOnlyInfo("gRPC server for account microservice registered and starting...")
	return server.Serve(lis)
}

func (g *grpcServer) PostAccount(ctx context.Context, req *pb.PostAccountRequest) (*pb.PostAccountResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Received PostAccount gRPC request for name: " + req.GetName())

	acc, err := g.service.PostAccount(ctx, req.GetName(), req.GetEmail(), req.GetPasswordHash(), req.GetRole())
	if err != nil {
		Logs.Error(ctx, "PostAccount service error: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "New account created with ID: "+acc.ID)

	//sending account created email
	// sending account created email using template
	// err = g.mailClient.SendEmail(ctx, acc.Email, "Welcome to Zenvis!", "account_created", map[string]string{
	// 	"Name":  acc.Name,
	// 	// "Email": acc.Email,
	// })
	// if err != nil {
	// 	Logs.Error(ctx, "Failed to send welcome email: "+err.Error())
	// }

	//now using pub sub model
	emailJob := map[string]interface{}{
		"to":           acc.Email,
		"subject":      "Welcome to Zenvis!",
		"templateName": "account_created",
		"templateData": map[string]string{
			"Name":  acc.Name,
			"Email": acc.Email,
		},
	}

	payload, err := json.Marshal(emailJob)
	if err != nil {
		Logs.Error(ctx, "Failed to marshal email job: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Email job publishing to NATS")
	err = g.netScan.Publish("emails.send", payload)
	if err != nil {
		Logs.Error(ctx, "Failed to publish email job: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Email job published to NATS")

	return &pb.PostAccountResponse{
		Account: &pb.Account{
			Id:    acc.ID,
			Name:  acc.Name,
			Email: acc.Email,
			Role: acc.Role,
			TokenVersion: acc.TokenVersion,
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
			Id:    acc.ID,
			Name:  acc.Name,
			Email: acc.Email,
			Role: acc.Role,
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
			Id:    acc.ID,
			Name:  acc.Name,
			Email: acc.Email,
			Role: acc.Role,
		}
	}

	return &pb.GetAccountsResponse{
		Accounts: grpcAccs,
	}, nil
}

func (g *grpcServer) GetEmail(ctx context.Context, req *pb.GetEmailRequest) (*pb.GetEmailResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Received GetEmail gRPC request for name: " + req.GetName())

	email, err := g.service.GetEmail(ctx, req.GetName())
	if err != nil {
		Logs.Error(ctx, "GetEmail service error: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Fetched email: "+email)

	return &pb.GetEmailResponse{
		Email: email,
	}, nil
}

func (g *grpcServer) GetEmailForAuth(ctx context.Context, req *pb.GetEmailForAuthRequest) (*pb.GetEmailForAuthResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Received GetEmailForAuth gRPC request for email: " + req.GetEmail())

	email, err := g.service.GetEmailForAuth(ctx, req.GetEmail())
	if err != nil {
		Logs.Error(ctx, "GetEmailForAuth service error: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Fetched email: "+email.Email)

	return &pb.GetEmailForAuthResponse{
		Id: email.ID,
		Name: email.Name,
		Email: email.Email,
		PasswordHash: email.PasswordHash,
		Role: email.Role,
		TokenVersion: email.TokenVersion,
	}, nil
}

func (g *grpcServer) IncrementTokenVersion(ctx context.Context, req *pb.IncrementTokenVersionRequest) (*pb.IncrementTokenVersionResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Received IncrementTokenVersion gRPC request for user ID: " + req.GetUserId())

	if err := g.service.IncrementTokenVersion(ctx, req.GetUserId()); err != nil {
		Logs.Error(ctx, "IncrementTokenVersion service error: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Incremented token version for user ID: "+req.GetUserId())

	return &pb.IncrementTokenVersionResponse{Ok: true}, nil
}
