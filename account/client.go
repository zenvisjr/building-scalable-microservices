package account

import (
	"context"
	"fmt"

	"github.com/zenvisjr/building-scalable-microservices/account/pb"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// var logs = logger.GetGlobalLogger()

type Client struct {
	conn    *grpc.ClientConn
	service pb.AccountServiceClient
}

// Connects to gRPC server and initializes service client
func NewClient(address string) (*Client, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Connecting to Account gRPC service at " + address)

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		Logs.Warn(context.Background(), "Failed to connect to Account gRPC: "+err.Error())
		return nil, err
	}

	Logs.LocalOnlyInfo("Connected to Account gRPC service")
	service := pb.NewAccountServiceClient(conn)

	return &Client{
		conn:    conn,
		service: service,
	}, nil
}

func (c *Client) Close() {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Closing gRPC connection to Account service")
	c.conn.Close()
}

func (c *Client) PostAccount(ctx context.Context, name string, email string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Sending PostAccount request for name: " + name)

	resp, err := c.service.PostAccount(ctx, &pb.PostAccountRequest{
		Name: name,
		Email: email,
	})
	if err != nil {
		Logs.Error(ctx, "PostAccount RPC failed: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Account created with ID: "+resp.Account.Id)
	return &Account{
		ID:   resp.Account.Id,
		Name: resp.Account.Name,
		Email: resp.Account.Email,
	}, nil
}

func (c *Client) GetAccount(ctx context.Context, id string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Fetching account with ID: " + id)

	resp, err := c.service.GetAccount(ctx, &pb.GetAccountRequest{Id: id})
	if err != nil {
		Logs.Error(ctx, "GetAccount RPC failed: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Fetched account: "+resp.Account.Name)
	return &Account{
		ID:   resp.Account.Id,
		Name: resp.Account.Name,
		Email: resp.Account.Email,
		}, nil
}

func (c *Client) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo(fmt.Sprintf("Fetching accounts with pagination: skip=%d take=%d", skip, take))

	resp, err := c.service.GetAccounts(ctx, &pb.GetAccountsRequest{Skip: skip, Take: take})
	if err != nil {
		Logs.Error(ctx, "GetAccounts RPC failed: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, fmt.Sprintf("Fetched %d accounts", len(resp.Accounts)))

	accounts := make([]Account, len(resp.Accounts))
	for i, acc := range resp.Accounts {
		accounts[i] = Account{
			ID:   acc.Id,
			Name: acc.Name,
			Email: acc.Email,
		}
	}
	return accounts, nil
}

func (c *Client) GetEmail(ctx context.Context, name string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Fetching email for account with name: " + name)

	resp, err := c.service.GetEmail(ctx, &pb.GetEmailRequest{Name: name})
	if err != nil {
		Logs.Error(ctx, "GetEmail RPC failed: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Fetched email: "+resp.Email)
	return &Account{
		Email: resp.Email,
	}, nil
}
