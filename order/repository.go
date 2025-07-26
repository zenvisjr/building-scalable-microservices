package order

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/zenvisjr/building-scalable-microservices/logger"
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
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Initializing PostgreSQL repository...")

	db, err := sql.Open("postgres", url)
	if err != nil {
		Logs.Error(context.Background(), "Failed to open DB connection: "+err.Error())
		return nil, err
	}

	if err := db.Ping(); err != nil {
		Logs.Error(context.Background(), "Failed to ping the database: "+err.Error())
		return nil, err
	}

	Logs.Info(context.Background(), "Successfully connected to PostgreSQL")
	return &postgresRepository{db}, nil
}

func (p *postgresRepository) Close() {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Closing PostgreSQL connection")
	p.db.Close()
}
func (p *postgresRepository) PutOrder(ctx context.Context, order Order) (err error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo(fmt.Sprintf("Inserting order ID: %s with %d products", order.ID, len(order.Products)))

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		Logs.Error(ctx, "Failed to begin transaction: "+err.Error())
		return err
	}

	defer func() {
		if err != nil {
			Logs.Error(ctx, "Transaction rollback due to error: "+err.Error())
			_ = tx.Rollback()
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				Logs.Error(ctx, "Transaction commit failed: "+commitErr.Error())
				err = commitErr
			} else {
				Logs.Info(ctx, fmt.Sprintf("Successfully committed order ID: %s", order.ID))
			}
		}
	}()

	// Insert order metadata
	_, err = tx.ExecContext(ctx,
		"INSERT INTO orders(id, created_at, account_id, total_price) VALUES($1, $2, $3, $4)",
		order.ID, order.CreatedAt, order.AccountID, order.TotalPrice)
	if err != nil {
		Logs.Error(ctx, "Failed to insert order record: "+err.Error())
		return err
	}
	Logs.LocalOnlyInfo("Inserted order metadata")

	// Prepare COPY statement
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("order_products", "order_id", "product_id", "quantity"))
	if err != nil {
		Logs.Error(ctx, "Failed to prepare COPY statement: "+err.Error())
		return err
	}
	defer stmt.Close()

	for _, p := range order.Products {
		Logs.LocalOnlyInfo(fmt.Sprintf("Inserting product ID: %s with quantity %d", p.ProductID, p.Quantity))
		_, err = stmt.ExecContext(ctx, order.ID, p.ProductID, p.Quantity)
		if err != nil {
			Logs.Error(ctx, "Failed to insert product into COPY buffer: "+err.Error())
			return err
		}
	}

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		Logs.Error(ctx, "COPY statement finalization failed: "+err.Error())
		return err
	}

	return nil
}
func (p *postgresRepository) ListOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Fetching orders from DB for account: " + accountID)

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
		Logs.Error(ctx, "Failed to query orders from DB: "+err.Error())
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
			accID      string
			totalPrice float64
			productID  string
			quantity   uint32
		)

		if err := rows.Scan(&orderID, &createdAt, &accID, &totalPrice, &productID, &quantity); err != nil {
			Logs.Error(ctx, "Failed to scan order row: "+err.Error())
			return nil, err
		}

		if lastOrderID != orderID {
			if currentOrder != nil {
				orders = append(orders, *currentOrder)
			}
			Logs.LocalOnlyInfo("New order ID encountered: " + orderID)

			currentOrder = &Order{
				ID:         orderID,
				CreatedAt:  createdAt,
				AccountID:  accID,
				TotalPrice: totalPrice,
				Products:   []OrderedProduct{},
			}
			lastOrderID = orderID
		}

		Logs.LocalOnlyInfo(fmt.Sprintf("Adding product %s (qty: %d) to order %s", productID, quantity, orderID))

		currentOrder.Products = append(currentOrder.Products, OrderedProduct{
			ProductID: productID,
			Quantity:  quantity,
		})
	}

	if currentOrder != nil {
		orders = append(orders, *currentOrder)
	}

	if err = rows.Err(); err != nil {
		Logs.Error(ctx, "Error after iterating rows: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, fmt.Sprintf("Returning %d orders for account: %s", len(orders), accountID))
	return orders, nil
}
