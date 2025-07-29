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
	GetEmailByName(ctx context.Context, name string) (string, error)
	GetAccountForAuth(ctx context.Context, email string) (*Account, error)
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

	_, err := p.db.ExecContext(ctx, "INSERT INTO accounts(id, name, email, password_hash, role) VALUES($1, $2, $3, $4, $5)", acc.ID, acc.Name, acc.Email, acc.PasswordHash, acc.Role)
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
	row := p.db.QueryRowContext(ctx, "SELECT id, name, email, role FROM accounts WHERE id = $1", id)
	a := &Account{}
	err := row.Scan(&a.ID, &a.Name, &a.Email, &a.Role)
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

	rows, err := p.db.QueryContext(ctx, "SELECT id, name, email, role FROM accounts ORDER BY id DESC LIMIT $1 OFFSET $2", limit, skip)
	if err != nil {
		Logs.Error(ctx, "Failed to list accounts: "+err.Error())
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		a := &Account{}
		if err := rows.Scan(&a.ID, &a.Name, &a.Email, &a.Role); err != nil {
			Logs.Error(ctx, "Failed to scan account row: "+err.Error())
			return nil, err
		}
		accounts = append(accounts, *a)
	}

	Logs.Info(ctx, "Returned " + logger.IntToStr(len(accounts)) + " accounts from DB")
	return accounts, nil
}

//used by email service
func (p *postgresRepository) GetEmailByName(ctx context.Context, name string) (string, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Fetching email for account with name: " + name)

	row := p.db.QueryRowContext(ctx, "SELECT email FROM accounts WHERE name = $1", name)
	email := ""
	err := row.Scan(&email)
	if err != nil {
		Logs.Error(ctx, "Email fetch failed: "+err.Error())
		return "", err
	}

	Logs.Info(ctx, "Fetched email: "+email)
	return email, nil
}

//used by auth service
func (p *postgresRepository) GetAccountForAuth(ctx context.Context, email string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Fetching account for auth with email: " + email)
	row := p.db.QueryRowContext(ctx, "SELECT id, name, email, password_hash, role FROM accounts WHERE email = $1", email)
	a := &Account{}
	err := row.Scan(&a.ID, &a.Name, &a.Email, &a.PasswordHash, &a.Role)
	if err != nil {
		Logs.Error(ctx, "Account fetch failed: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Fetched account for auth with ID: "+a.ID)
	return a, nil
}

