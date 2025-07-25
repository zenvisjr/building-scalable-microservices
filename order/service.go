package order

import (
	"context"
	"time"

	"github.com/segmentio/ksuid"
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
}


type orderService struct {
	repo Repository
}

func NewOrderService(r Repository) (Service, error) {
	return &orderService{r}, nil
}

func (o *orderService) PostOrder(ctx context.Context, accountID string, orders []OrderedProduct) (*Order, error) {
	order := &Order{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
		AccountID: accountID,
		Products:  orders,
	}
	totalPrice := 0.0
	for _, placedOrder := range orders {
		totalPrice += placedOrder.Price * float64(placedOrder.Quantity)
	}
	order.TotalPrice = totalPrice

	err := o.repo.PutOrder(ctx, *order)
	if err != nil {
		return nil, err
	}
	return order, nil

}

func (o *orderService) GetOrdersByAccount(ctx context.Context, accountID string) ([]Order, error) {
	// order := &Or
	// return  o.repo.PutOrder(ctx, order)
	return o.repo.ListOrdersForAccount(ctx, accountID)

}
