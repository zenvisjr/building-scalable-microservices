package order

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type Service interface {
	PostOrder(ctx context.Context, accountID string, orders []OrderedProduct) (*Order, error)
	GetOrdersByAccount(ctx context.Context, accountid string) ([]Order, error)
}

type Order struct {
	ID         string           `json:"id"`
	CreatedAt  time.Time        `json:"created_at"`
	AccountID  string           `json:"account_id"`
	TotalPrice float64          `json:"price"`
	Products   []OrderedProduct `json:"orderedproducts"`
}

type OrderedProduct struct {
	ProductID   string  `json:"productid"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    uint32  `json:"quantity"`
	Stock       uint32  `json:"stock"`
}

type orderService struct {
	repo Repository
}

func NewOrderService(r Repository) (Service, error) {
	return &orderService{r}, nil
}

func (o *orderService) PostOrder(ctx context.Context, accountID string, orders []OrderedProduct) (*Order, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo(fmt.Sprintf("PostOrder called with %d products", len(orders)))

	order := &Order{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
		AccountID: accountID,
		Products:  orders,
	}

	order.TotalPrice = 0.0
	for _, placedOrder := range orders {
		order.TotalPrice += placedOrder.Price * float64(placedOrder.Quantity)
	}

	Logs.LocalOnlyInfo(fmt.Sprintf("Prepared order: TotalPrice=%.2f, Products=%d", order.TotalPrice, len(order.Products)))

	err := o.repo.CreateOrder(ctx, *order)
	if err != nil {
		Logs.Error(ctx, "Failed to save order: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, fmt.Sprintf("Order saved: ID=%s, TotalPrice=%.2f", order.ID, order.TotalPrice))
	return order, nil
}

func (o *orderService) GetOrdersByAccount(ctx context.Context, accountID string) ([]Order, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Fetching orders for account: "+accountID)

	orders, err := o.repo.ListOrdersForAccount(ctx, accountID)
	if err != nil {
		Logs.Error(ctx, "Failed to fetch orders: "+err.Error())
		return nil, err
	}

	Logs.LocalOnlyInfo(fmt.Sprintf("Fetched %d orders for account %s", len(orders), accountID))
	return orders, nil
}
