package account

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

// var Logs = logger.GetGlobalLogger()

type Repository interface {
	Close()
	PutAccount(ctx context.Context, acc Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, skip uint64, limit uint64) ([]Account, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Opening PostgreSQL connection")

	db, err := sql.Open("postgres", url)
	if err != nil {
		Logs.Error(context.Background(), "Failed to open DB: "+err.Error())
		return nil, err
	}

	if err := db.Ping(); err != nil {
		Logs.Error(context.Background(), "Failed to ping DB: "+err.Error())
		return nil, err
	}

	Logs.Info(context.Background(), "Successfully connected to PostgreSQL")
	return &postgresRepository{db: db}, nil
}

func (p *postgresRepository) Close() {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Closing DB connection")
	p.db.Close()
}

func (p *postgresRepository) Ping() error {
	return p.db.Ping()
}

func (p *postgresRepository) PutAccount(ctx context.Context, acc Account) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Inserting new account into DB")

	_, err := p.db.ExecContext(ctx, "INSERT INTO accounts(id, name) VALUES($1, $2)", acc.ID, acc.Name)
	if err != nil {
		Logs.Error(ctx, "Failed to insert account: "+err.Error())
		return err
	}

	Logs.Info(ctx, "Inserted account with ID: "+acc.ID)
	return nil
}

func (p *postgresRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Fetching account by ID: " + id)
	row := p.db.QueryRowContext(ctx, "SELECT id, name FROM accounts WHERE id = $1", id)
	a := &Account{}
	err := row.Scan(&a.ID, &a.Name)
	if err != nil {
		Logs.Error(ctx, "Account fetch failed: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Fetched account with ID: "+a.ID)
	return a, nil
}

func (p *postgresRepository) ListAccounts(ctx context.Context, skip uint64, limit uint64) ([]Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Listing accounts with limit and skip")

	rows, err := p.db.QueryContext(ctx, "SELECT id, name FROM accounts ORDER BY id DESC LIMIT $1 OFFSET $2", limit, skip)
	if err != nil {
		Logs.Error(ctx, "Failed to list accounts: "+err.Error())
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		a := &Account{}
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			Logs.Error(ctx, "Failed to scan account row: "+err.Error())
			return nil, err
		}
		accounts = append(accounts, *a)
	}

	Logs.Info(ctx, "Returned " + logger.IntToStr(len(accounts)) + " accounts from DB")
	return accounts, nil
}
