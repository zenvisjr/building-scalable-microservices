package order

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, order Order) error
	ListOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		fmt.Println("failed to ping the db")
		return nil, err
	}
	return &postgresRepository{db}, nil
}

func (p *postgresRepository) Close() {
	p.db.Close()
}

func (p *postgresRepository) PutOrder(ctx context.Context, order Order) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		} else {
			err = tx.Commit()
		}
	}()

	// FIRST: Insert the order record
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO orders(id, created_at, account_id, total_price) VALUES($1, $2, $3, $4)",
		order.ID, order.CreatedAt, order.AccountID, order.TotalPrice)

	if err != nil {
		return err
	}

	// SECOND: Insert order products using COPY
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("order_products", "order_id", "product_id", "quantity"))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range order.Products {
		_, err = stmt.ExecContext(ctx, order.ID, p.ProductID, p.Quantity)
		if err != nil {
			return err
		}
	}
	
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}

	return nil
}
func (p *postgresRepository) ListOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT 
			o.id, o.created_at, o.account_id, 
			o.total_price::money::numeric::float8, 
			op.product_id, op.quantity
		FROM orders o
		JOIN order_products op ON o.id = op.order_id
		WHERE o.account_id = $1
		ORDER BY o.id`,
		accountID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		currentOrder *Order
		lastOrderID  string
		orders       []Order
	)

	for rows.Next() {
		var (
			orderID    string
			createdAt  time.Time
			accountID  string
			totalPrice float64
			productID  string
			quantity   uint32
		)

		if err := rows.Scan(&orderID, &createdAt, &accountID, &totalPrice, &productID, &quantity); err != nil {
			return nil, err
		}

		if lastOrderID != orderID {
			if currentOrder != nil {
				orders = append(orders, *currentOrder)
			}

			currentOrder = &Order{
				ID:         orderID,
				CreatedAt:  createdAt,
				AccountID:  accountID,
				TotalPrice: totalPrice,
				Products:   []OrderedProduct{},
			}
			lastOrderID = orderID
		}

		currentOrder.Products = append(currentOrder.Products, OrderedProduct{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	if currentOrder != nil {
		orders = append(orders, *currentOrder)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
