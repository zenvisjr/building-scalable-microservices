package account

import (
	"context"

	"github.com/segmentio/ksuid"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role"`
	TokenVersion int32  `json:"token_version"`
	IsActive     bool   `json:"is_active"`
}

type Service interface {
	PostAccount(ctx context.Context, name string, email string, passwordHash string, role string) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
	GetEmail(ctx context.Context, id string) (string, error)
	GetEmailForAuth(ctx context.Context, email string) (*Account, error)
	IncrementTokenVersion(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, email string, password string) error
	DeactivateAccount(ctx context.Context, userID string) error
	ReactivateAccount(ctx context.Context, userID string) error
	DeleteAccount(ctx context.Context, userID string) error
}

type accountService struct {
	repo Repository
}

func NewAccountService(r Repository) Service {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Creating new AccountService instance")
	return &accountService{repo: r}
}

func HashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashBytes), nil
}
func (a *accountService) PostAccount(ctx context.Context, name string, email, plainPassword, role string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("PostAccount called with name: " + name)

	// Hash the password
	Logs.LocalOnlyInfo("Hashing password for account: " + name)
	hashedPassword, err := HashPassword(plainPassword)
	if err != nil {
		Logs.Error(ctx, "Password hashing failed: "+err.Error())
		return nil, err
	}

	account := &Account{
		ID:           ksuid.New().String(),
		Name:         name,
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
		TokenVersion: 1,
	}
	err = a.repo.CreateAccount(ctx, *account)
	if err != nil {
		Logs.Error(ctx, "Failed to store new account: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "New account created with ID: "+account.ID)
	return account, nil
}

func (a *accountService) GetAccount(ctx context.Context, id string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("GetAccount called with ID: " + id)

	account, err := a.repo.GetAccountByID(ctx, id)
	if err != nil {
		Logs.Error(ctx, "Failed to get account by ID: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Fetched account with ID: "+account.ID)
	return account, nil
}

func (a *accountService) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("GetAccounts called with skip=" +
		logger.Uint64ToStr(skip) + ", take=" +
		logger.Uint64ToStr(take))

	if skip > 100 || (take == 0 && skip == 0) {
		take = 100
	}

	accounts, err := a.repo.ListAccounts(ctx, skip, take)
	if err != nil {
		Logs.Error(ctx, "Failed to list accounts: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Fetched accounts count: "+logger.IntToStr(len(accounts)))
	return accounts, nil
}

func (a *accountService) GetEmail(ctx context.Context, name string) (string, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("GetEmail called with name: " + name)

	email, err := a.repo.GetEmailByName(ctx, name)
	if err != nil {
		Logs.Error(ctx, "Failed to get email by name: "+err.Error())
		return "", err
	}

	Logs.Info(ctx, "Fetched email: "+email)
	return email, nil
}

func (a *accountService) GetEmailForAuth(ctx context.Context, email string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("GetEmailForAuth called with email: " + email)

	account, err := a.repo.GetAccountForAuth(ctx, email)
	if err != nil {
		Logs.Error(ctx, "Failed to get account for auth: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Fetched account for auth with ID: "+account.ID)
	return account, nil
}

func (a *accountService) IncrementTokenVersion(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Incrementing token version for user: " + userID)

	return a.repo.IncrementTokenVersion(ctx, userID)
}

func (a *accountService) UpdatePassword(ctx context.Context, email string, password string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Updating password for email in service: " + email)

	// Hash the password
	Logs.LocalOnlyInfo("Hashing password for account: " + email)
	hashedPassword, err := HashPassword(password)
	Logs.LocalOnlyInfo("hashed password: " + hashedPassword)
	if err != nil {
		Logs.Error(ctx, "Password hashing failed: "+err.Error())
		return err
	}
	if err := a.repo.UpdatePassword(ctx, email, hashedPassword); err != nil {
		Logs.Error(ctx, "Failed to update password: "+err.Error())
		return err
	}
	Logs.Info(ctx, "Updated password for email in service: "+email)
	return nil
}

func (a *accountService) DeactivateAccount(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Got deactivate account request in service for user: " + userID)

	return a.repo.DeactivateAccount(ctx, userID)
}

func (a *accountService) ReactivateAccount(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Got reactivate account request in service for user: " + userID)

	return a.repo.ReactivateAccount(ctx, userID)
}


func (a *accountService) DeleteAccount(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Got delete account request in service for user: " + userID)

	return a.repo.DeleteAccount(ctx, userID)
}


