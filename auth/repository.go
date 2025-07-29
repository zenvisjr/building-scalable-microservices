package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type Repository interface {
	StoreRefreshToken(ctx context.Context, refreshToken string, userId string) error
	GetRefreshToken(ctx context.Context, id string) (*TokenData, error)
	DeleteRefreshToken(ctx context.Context, id string) error
	Close()
	Ping() error
}

type postgresRepository struct {
	db *sql.DB
}

type TokenData struct {
	UserID       string
	RefreshToken string
	ExpiresAt    time.Time
}

func NewPostgresRepository(url string) (Repository, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Opening access_token PostgreSQL connection")

	db, err := sql.Open("postgres", url)
	if err != nil {
		Logs.Error(context.Background(), "Failed to open DB: "+err.Error())
		return nil, err
	}

	if err := db.Ping(); err != nil {
		Logs.Error(context.Background(), "Failed to ping DB: "+err.Error())
		return nil, err
	}

	Logs.Info(context.Background(), "Successfully connected to access_token PostgreSQL")
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

func (p *postgresRepository) StoreRefreshToken(ctx context.Context, refreshToken string, userId string) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Storing refresh token in DB")
	query := `
		INSERT INTO refresh_tokens (user_id, refresh_token, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := p.db.ExecContext(ctx, query, userId, refreshToken, time.Now().Add(15*time.Minute))
	return err
}

func (r *postgresRepository) GetRefreshToken(ctx context.Context, id string) (*TokenData, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Getting refresh token from access_token DB")
	query := `
		SELECT user_id, refresh_token, expires_at
		FROM refresh_tokens
		WHERE user_id = $1
		ORDER BY expires_at DESC
		LIMIT 1;
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var rt TokenData
	if err := row.Scan(&rt.UserID, &rt.RefreshToken, &rt.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}
	return &rt, nil
}

func (r *postgresRepository) DeleteRefreshToken(ctx context.Context, userID string) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}


