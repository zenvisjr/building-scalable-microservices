package account

import (
	"context"

	"github.com/segmentio/ksuid"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)



type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

type Service interface {
	PostAccount(ctx context.Context, name string, email string) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
	GetEmail(ctx context.Context, id string) (string, error)
}

type accountService struct {
	repo Repository
}

func NewAccountService(r Repository) Service {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Creating new AccountService instance")
	return &accountService{repo: r}
}

func (a *accountService) PostAccount(ctx context.Context, name string, email string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("PostAccount called with name: " + name)

	account := &Account{
		ID:   ksuid.New().String(),
		Name: name,
		Email: email,
	}
	err := a.repo.PutAccount(ctx, *account)
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

	Logs.Info(ctx, "Fetched accounts count: " + logger.IntToStr(len(accounts)))
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
	