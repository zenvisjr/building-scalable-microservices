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

func (c *Client) PostAccount(ctx context.Context, name, email, plainPassword, role string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Sending PostAccount request for name: " + name)

	resp, err := c.service.PostAccount(ctx, &pb.PostAccountRequest{
		Name: name,
		Email: email,
		PasswordHash: plainPassword,
		Role: role,
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
		Role: resp.Account.Role,
		TokenVersion: resp.Account.TokenVersion,
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
		Role: resp.Account.Role,
		IsActive: resp.Account.IsActive,
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
			Role: acc.Role,
			IsActive: acc.IsActive,
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


func (c *Client) GetEmailForAuth(ctx context.Context, email string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Fetching email for account with email: " + email)

	resp, err := c.service.GetEmailForAuth(ctx, &pb.GetEmailForAuthRequest{Email: email})
	if err != nil {
		Logs.Error(ctx, "GetEmailForAuth RPC failed: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Fetched email: "+resp.Email)
	return &Account{
		ID: resp.Id,
		Name: resp.Name,
		Email: resp.Email,
		PasswordHash: resp.PasswordHash,
		Role: resp.Role,
		TokenVersion: resp.TokenVersion,
		IsActive: resp.IsActive,
	}, nil
}

func (c *Client) IncrementTokenVersion(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Incrementing token version for user ID: " + userID)

	_, err := c.service.IncrementTokenVersion(ctx, &pb.IncrementTokenVersionRequest{UserId: userID})
	if err != nil {
		Logs.Error(ctx, "IncrementTokenVersion RPC failed: "+err.Error())
		return err
	}
	Logs.Info(ctx, "Incremented token version for user ID: "+userID)
	return nil
}

func(c *Client) UpdatePassword(ctx context.Context, email string, password string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Updating password for email in client: " + email)

	_, err := c.service.UpdatePassword(ctx, &pb.UpdatePasswordRequest{Email: email, Password: password})
	if err != nil {
		Logs.Error(ctx, "UpdatePassword RPC failed: "+err.Error())
		return err
	}
	Logs.Info(ctx, "Updated password for email in client: "+email)
	return nil
}

func(c *Client) DeactivateAccount(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Deactivating account for user ID: " + userID)

	_, err := c.service.DeactivateAccount(ctx, &pb.UpdateAccountRequest{UserId: userID})
	if err != nil {
		Logs.Error(ctx, "DeactivateAccount RPC failed: "+err.Error())
		return err
	}
	Logs.Info(ctx, "Deactivated account for user ID: "+userID)
	return nil
}

func(c *Client) ReactivateAccount(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Reactivating account for user ID: " + userID)

	_, err := c.service.ReactivateAccount(ctx, &pb.UpdateAccountRequest{UserId: userID})
	if err != nil {
		Logs.Error(ctx, "ReactivateAccount RPC failed: "+err.Error())
		return err
	}
	Logs.Info(ctx, "Reactivated account for user ID: "+userID)
	return nil
}



func(c *Client) DeleteAccount(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Deleting account for user ID: " + userID)

	_, err := c.service.DeleteAccount(ctx, &pb.UpdateAccountRequest{UserId: userID})
	if err != nil {
		Logs.Error(ctx, "DeleteAccount RPC failed: "+err.Error())
		return err
	}
	Logs.Info(ctx, "Deleted account for user ID: "+userID)
	return nil
}