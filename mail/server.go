package mail

import (
	"context"
	"fmt"
	"net"

	"github.com/zenvisjr/building-scalable-microservices/logger"
	"github.com/zenvisjr/building-scalable-microservices/mail/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type MailServer struct {
	service Service
	pb.UnimplementedMailServiceServer
}


func InitMailGRPC(service Service, port int) error {
	Logs := logger.GetGlobalLogger()
	address := fmt.Sprintf(":%d", port)
	Logs.LocalOnlyInfo("Initializing gRPC listener for Mail microservice on " + address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		Logs.Error(context.Background(), "Failed to bind to port: "+err.Error())
		return err
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryLoggingInterceptor()),
	)

	pb.RegisterMailServiceServer(server, &MailServer{service: service})
	reflection.Register(server)

	Logs.LocalOnlyInfo("gRPC server for Mail microservice registered and starting...")
	return server.Serve(lis)
}

func(g *MailServer) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Received SendEmail gRPC request for email: " + req.GetTo())

	err := g.service.SendEmail(ctx, req.GetTo(), req.GetSubject(), req.GetTemplateName(), req.GetTemplateData())
	if err != nil {
		Logs.Error(ctx, "SendEmail service error: "+err.Error())
		return nil, err
	}


	Logs.Info(ctx, "Email sent to "+req.GetTo())
	return &pb.SendEmailResponse{Ok: true}, nil
}
