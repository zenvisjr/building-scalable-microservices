package account

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

// var Logs = logger.GetGlobalLogger()

var (
	errAccountDeactivated = fmt.Errorf("Account is not active")
)

type Repository interface {
	Close()
	PutAccount(ctx context.Context, acc Account) error
	GetAccountByID(ctx context.Context, id string) (*Account, error)
	ListAccounts(ctx context.Context, skip uint64, limit uint64) ([]Account, error)
	GetEmailByName(ctx context.Context, name string) (string, error)
	GetAccountForAuth(ctx context.Context, email string) (*Account, error)
	IncrementTokenVersion(ctx context.Context, userID string) error
	UpdatePassword(ctx context.Context, email string, password_hash string) error
	DeactivateAccount(ctx context.Context, userID string) error
	ReactivateAccount(ctx context.Context, userID string) error
	DeleteAccount(ctx context.Context, userID string) error
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

	_, err := p.db.ExecContext(ctx,
		"INSERT INTO accounts(id, name, email, password_hash, role, token_version) VALUES($1, $2, $3, $4, $5, $6)",
		acc.ID, acc.Name, acc.Email, acc.PasswordHash, acc.Role, acc.TokenVersion)
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
	row := p.db.QueryRowContext(ctx, "SELECT id, name, email, role, is_active FROM accounts WHERE id = $1", id)
	a := &Account{}
	err := row.Scan(&a.ID, &a.Name, &a.Email, &a.Role, &a.IsActive)
	if err != nil {
		Logs.Error(ctx, "Account fetch failed: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Fetched account with ID: "+a.ID)
	return a, nil
}

func (p *postgresRepository) ListAccounts(ctx context.Context, skip uint64, limit uint64) ([]Account, error) {
	Logs := logger.GetGlobalLogger()
	// Logs.LocalOnlyInfo("Listing accounts with limit and skip")

	rows, err := p.db.QueryContext(ctx, "SELECT id, name, email, role, is_active FROM accounts ORDER BY id DESC LIMIT $1 OFFSET $2", limit, skip)
	if err != nil {
		Logs.Error(ctx, "Failed to list accounts: "+err.Error())
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		a := &Account{}
		if err := rows.Scan(&a.ID, &a.Name, &a.Email, &a.Role, &a.IsActive); err != nil {
			Logs.Error(ctx, "Failed to scan account row: "+err.Error())
			return nil, err
		}
		accounts = append(accounts, *a)
	}

	Logs.Info(ctx, "Returned "+logger.IntToStr(len(accounts))+" accounts from DB")
	return accounts, nil
}

// used by email service
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

// used by auth service
func (p *postgresRepository) GetAccountForAuth(ctx context.Context, email string) (*Account, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Fetching account for auth with email: " + email)
	row := p.db.QueryRowContext(ctx, "SELECT id, name, email, password_hash, role, token_version, is_active FROM accounts WHERE email = $1", email)
	a := &Account{}
	err := row.Scan(&a.ID, &a.Name, &a.Email, &a.PasswordHash, &a.Role, &a.TokenVersion, &a.IsActive)
	if err != nil {
		Logs.Error(ctx, "Account fetch failed: "+err.Error())
		return nil, err
	}
	// if !a.IsActive {
	// 	Logs.Error(ctx, "Account is not active")
	// 	return nil, errAccountDeactivated
	// }

	Logs.Info(ctx, "Fetched account for auth with ID: "+a.ID)
	return a, nil
}

func (r *postgresRepository) IncrementTokenVersion(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Incrementing token version in DB")
	query := `UPDATE accounts SET token_version = token_version + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (p *postgresRepository) UpdatePassword(ctx context.Context, email string, password_hash string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Updating password for email in DB: " + email)

	query := `UPDATE accounts SET password_hash = $1 WHERE email = $2`
	res, err := p.db.ExecContext(ctx, query, password_hash, email)
	count, _ := res.RowsAffected()
	Logs.LocalOnlyInfo(fmt.Sprintf("Password update affected %d row(s)", count))

	if err != nil {
		Logs.Error(ctx, "Password update failed: "+err.Error())
		return err
	}

	return nil
}

func (p *postgresRepository) DeactivateAccount(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Deactivating account for user: " + userID)

	query := `UPDATE accounts SET is_active = false WHERE id = $1`
	_, err := p.db.ExecContext(ctx, query, userID)
	return err
}

func (p *postgresRepository) ReactivateAccount(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Reactivating account for user: " + userID)

	query := `UPDATE accounts SET is_active = true WHERE id = $1`
	_, err := p.db.ExecContext(ctx, query, userID)
	return err
}

func (p *postgresRepository) DeleteAccount(ctx context.Context, userID string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Deleting account for user: " + userID)

	query := `DELETE FROM accounts WHERE id = $1`
	_, err := p.db.ExecContext(ctx, query, userID)
	return err
}