package account

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
)

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
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{
		db: db,
	}, nil
}

func(p *postgresRepository) Close() {
	p.db.Close()
}

func(p *postgresRepository) Ping() error {
	return p.db.Ping()
}

func(p *postgresRepository) PutAccount(ctx context.Context, acc Account) error {
	_, err := p.db.ExecContext(ctx, "INSERT INTO accounts(id, name) VALUES($1, $2)", acc.ID, acc.Name)
	return err
}

func(p *postgresRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	row := p.db.QueryRowContext(ctx, "SELECT id, name FROM accounts WHERE id = $1", id)
	a := &Account{}
	err := row.Scan(a.ID, a.Name)
	if err != nil { 
		return nil, err
	}
	return a, nil
}

func(p *postgresRepository) ListAccounts(ctx context.Context, skip uint64, limit uint64) ([]Account, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, name FROM accounts ORDER BY id DESC LIMIT $1 OFFSET $2", limit, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		a := &Account{}
		if err := rows.Scan(a.ID, a.Name); err != nil {
			return nil, err
		}
		accounts = append(accounts, *a)
	}

	return accounts, nil
}
