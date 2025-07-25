package graphql

import (
	"context"
	"log"
	"time"
)

type accountResolver struct {
	server *Server
}

func (a *accountResolver) Orders(ctx context.Context, obj *Account) ([]*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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
		

	return orders, nil
}
