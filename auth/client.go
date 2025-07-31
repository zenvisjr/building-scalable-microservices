package auth

import (
	"context"
	"errors"

	"github.com/zenvisjr/building-scalable-microservices/auth/pb"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.AuthServiceClient
}

func NewClient(address string) (*Client, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Connecting to Auth gRPC service at " + address)

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		Logs.Error(context.Background(), "Failed to connect to Auth service: "+err.Error())
		return nil, err
	}

	Logs.LocalOnlyInfo("Connected to Auth gRPC service")
	return &Client{
		conn:    conn,
		service: pb.NewAuthServiceClient(conn),
	}, nil
}

func (c *Client) Close() {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Closing Auth service gRPC connection")
	c.conn.Close()
}

func (c *Client) Signup(ctx context.Context, name string, email string, password string, role string) (*pb.AuthResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Calling Auth gRPC service")
	return c.service.Signup(ctx, &pb.SignupRequest{
		Name:     name,
		Email:    email,
		Password: password,
		Role:     role,
	})
}

func (c *Client) Login(ctx context.Context, email string, password string) (*pb.AuthResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Calling Auth gRPC service")
	return c.service.Login(ctx, &pb.LoginRequest{
		Email:    email,
		Password: password,
	})
}

func (c *Client) RefreshToken(ctx context.Context, userId string) (*pb.AuthResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Calling Auth gRPC service")
	return c.service.RefreshToken(ctx, &pb.RefreshRequest{
		UserId: userId,
	})
}



type UserClaims struct {
	ID    string
	Email string
	Role  string
	TokenVersion int32
	// Name  string // optional
}

func (c *Client) VerifyToken(ctx context.Context, token string) (*UserClaims, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Verifying token in auth client...")
	resp, err := c.service.Verify(ctx, &pb.VerifyRequest{
		AccessToken: token,
	})
	if err != nil {
		Logs.Error(ctx, "Failed to verify token: "+err.Error())
		return nil, err
	}

	if resp == nil || resp.UserId == "" {
		Logs.Error(ctx, "Invalid token response")
		return nil, errors.New("invalid token response")
	}

	return &UserClaims{
		ID:    resp.UserId,
		Email: resp.Email,
		Role:  resp.Role,
				// Name: resp.Name, // only if your proto includes it
	}, nil
}

func (c *Client) Logout(ctx context.Context, userId string) (*pb.LogoutResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Calling Auth gRPC service for logout")
	resp, err := c.service.Logout(ctx, &pb.LogoutRequest{
		UserId: userId,
	})
	if err != nil {
		Logs.Error(ctx, "Failed to logout: "+err.Error())
		return nil, err
	}
	return resp, nil
}

func(c *Client) GetCurrent(ctx context.Context, skip uint64, take uint64, role string) (*pb.GetCurrentResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Calling Auth gRPC service for current users")
	resp, err := c.service.GetCurrent(ctx, &pb.GetCurrentRequest{
		Skip: skip,
		Take: take,
		Role: role,
	})
	if err != nil {
		Logs.Error(ctx, "Failed to get current users: "+err.Error())
		return nil, err
	}
	return resp, nil
}

func(c *Client) ResetPassword(ctx context.Context, email, password, userId string) (*pb.AuthResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Calling Auth gRPC service for reset password: " + email)
	resp, err := c.service.ResetPassword(ctx, &pb.ResetPasswordRequest{
		Email: email,
		Password: password,
		UserId: userId,
	})
	if err != nil {
		Logs.Error(ctx, "Failed to reset password: "+err.Error())
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeactivateAccount(ctx context.Context, userId string) (*pb.UpdateAccountResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Calling Auth gRPC service for deactivate account: " + userId)
	resp, err := c.service.DeactivateAccount(ctx, &pb.UpdateAccountRequest{
		UserId: userId,
	})
	if err != nil {
		Logs.Error(ctx, "Failed to deactivate account: "+err.Error())
		return nil, err
	}
	return resp, nil
}

func (c *Client) ReactivateAccount(ctx context.Context, userId string) (*pb.UpdateAccountResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Calling Auth gRPC service for reactivate account: " + userId)
	resp, err := c.service.ReactivateAccount(ctx, &pb.UpdateAccountRequest{
		UserId: userId,
	})
	if err != nil {
		Logs.Error(ctx, "Failed to reactivate account: "+err.Error())
		return nil, err
	}
	return resp, nil
}

func (c *Client) DeleteAccount(ctx context.Context, userId string) (*pb.UpdateAccountResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Calling Auth gRPC service for delete account: " + userId)
	resp, err := c.service.DeleteAccount(ctx, &pb.UpdateAccountRequest{
		UserId: userId,
	})
	if err != nil {
		Logs.Error(ctx, "Failed to delete account: "+err.Error())
		return nil, err
	}
	return resp, nil
}