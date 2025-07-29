package graphql

import (
	"context"
	"log"
	"time"

	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type accountResolver struct {
	server *Server
}

func (a *accountResolver) Orders(ctx context.Context, obj *Account) ([]*Order, error) {
	Logs := logger.GetGlobalLogger()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := RequireAdmin(ctx)
	if err != nil {
		Logs.Error(ctx, "No admin in incoming ctx before wrapping")
		return nil, err
	}

	Logs.Info(ctx, "Admin "+user.Email+" is fetching orders for account "+obj.Email)

	orderList, err := a.server.orderClient.GetOrdersForAccount(ctx, obj.ID)
	if err != nil {
		log.Panicln(err)
		return nil, err
	}
	var orders []*Order
	for _, order := range orderList {
		var orderedProduct []*OrderedProduct
		for _, product := range order.Products {
			orderedProduct = append(orderedProduct, &OrderedProduct{
				ID: product.ProductID,
				Name: product.Name,
				Description: product.Description,
				Price: product.Price,
				Quantity: int(product.Quantity),
			})
		}
		orders = append(orders, &Order{
			ID: order.ID,
			CreatedAt: order.CreatedAt.UTC().Format(time.RFC1123),
			TotalPrice: order.TotalPrice,
			Products: orderedProduct,
		})
	}
	Logs.Info(ctx, "Fetched orders for account "+obj.Email)

	return orders, nil
}
