package account

import (
	"context"
	"github.com/segmentio/ksuid"
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
	return &accountService{repo: r}
}

func (a *accountService) PostAccount(ctx context.Context, name string) (*Account, error) {
	account := &Account{
		ID:   ksuid.New().String(),
		Name: name,
	}
	err := a.repo.PutAccount(ctx, *account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (a *accountService) GetAccount(ctx context.Context, id string) (*Account, error) {
	return a.repo.GetAccountByID(ctx, id)
}

func (a *accountService) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	if skip > 100 || (take == 0 && skip == 100) {
		take = 100
	}
	
	return a.repo.ListAccounts(ctx, skip, take)
}
