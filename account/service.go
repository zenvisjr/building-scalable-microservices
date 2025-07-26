package account

import (
	"context"

	"github.com/segmentio/ksuid"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)



type Account struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Service interface {
	PostAccount(ctx context.Context, name string) (*Account, error)
	GetAccount(ctx context.Context, id string) (*Account, error)
	GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error)
}

type accountService struct {
	repo Repository
}

func NewAccountService(r Repository) Service {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Creating new AccountService instance")
	return &accountService{repo: r}
}

func (a *accountService) PostAccount(ctx context.Context, name string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("PostAccount called with name: " + name)

	account := &Account{
		ID:   ksuid.New().String(),
		Name: name,
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
