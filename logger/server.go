package logger

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/zenvisjr/building-scalable-microservices/logger/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	logFile *os.File
	pb.UnimplementedLoggerServiceServer
}

func ListenGRPC(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	lf, err := os.OpenFile("/root/centralized.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	fmt.Fprintf(lf, "[%s] Logger service started....\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(lf, "--------------------------------------------------------------------------\n")
	if err != nil {
		panic(err)
	}
	// defer lf.Close()
	server := grpc.NewServer()
	pb.RegisterLoggerServiceServer(server, &grpcServer{
		logFile: lf,
	})
	reflection.Register(server)
	return server.Serve(lis)
}

func (g *grpcServer) Log(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	// log.Println("Logger service received log request")
	line := fmt.Sprintf("[%s] [%s] [%s]", req.GetTimestamp(), req.GetService(), req.GetLevel())

	// if GetMethodFromContext(ctx) != "" {
	if req.GetMethod() != "" {
		// log.Println("we got a grpc intercept")
		line += fmt.Sprintf("  [%s]  ", req.GetMethod())
	}

	line += fmt.Sprintf(" %s\n", req.GetMsg())

	_, err := g.logFile.WriteString(line)
	if err != nil {
		return &pb.LogResponse{Ok: false}, err
	}

	// 🔥 Force flush buffer to disk
	err = g.logFile.Sync()
	if err != nil {
		return &pb.LogResponse{Ok: false}, err
	}

	// log.Println("Logger service logged request")
	return &pb.LogResponse{
		Ok: true,
	}, nil
}
