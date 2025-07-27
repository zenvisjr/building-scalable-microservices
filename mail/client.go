package mail

import (
	"context"

	"github.com/zenvisjr/building-scalable-microservices/logger"
	"github.com/zenvisjr/building-scalable-microservices/mail/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Mail struct {
	conn    *grpc.ClientConn
	service pb.MailServiceClient
}

func NewMailClient(address string) (*Mail, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Connecting to Mail gRPC service at " + address)

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		Logs.Warn(context.Background(), "Failed to connect to Mail gRPC: "+err.Error())
		return nil, err
	}

	Logs.LocalOnlyInfo("Connected to Mail gRPC service")
	service := pb.NewMailServiceClient(conn)

	return &Mail{
		conn:    conn,
		service: service,
	}, nil

}

func (c *Mail) Close() {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Closing gRPC connection to Mail service")
	c.conn.Close()
}

func (c *Mail) SendEmail(ctx context.Context, to, subject, templateName string, templateData map[string]string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Sending email to " + to)
	_, err := c.service.SendEmail(ctx, &pb.SendEmailRequest{
		To:           to,
		Subject:      subject,
		TemplateName: templateName,
		TemplateData: templateData,
	})
	if err != nil {
		Logs.Error(ctx, "SendEmail RPC failed: "+err.Error())
		return err
	}
	Logs.Info(ctx, "Email sent to "+to)
	return nil
}
