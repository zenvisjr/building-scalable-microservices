package order

import (
	"context"
	"log"

	"github.com/zenvisjr/building-scalable-microservices/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn *grpc.ClientConn
	service pb.OrderServiceClient
}


func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	service := pb.NewOrderServiceClient(conn)

	return &Client{
		conn: conn,
		service: service,
	}, nil
}

func(c *Client) Close() {
	c.conn.Close()
}

func(c *Client) PostOrder(ctx context.Context, id string, products []OrderedProduct) (*Order, error) {
	
	productList := []*pb.PostOrderRequest_OrderedProduct{}
	for _, p := range products {
		grpcProduct := &pb.PostOrderRequest_OrderedProduct{
			ProductId: p.ProductID,
			Quantity: p.Quantity,
		}
		productList = append(productList, grpcProduct)
	}

	grpcReq := &pb.PostOrderRequest{
		AccountId: id,
		Products: productList,
	}

	resp, err := c.service.PostOrder(ctx, grpcReq)
	if err != nil {
		log.Printf("PostOrder RPC failed: %v\n", err)
		return nil, err
	}
	newOrder := resp.Order

	return &Order{
		ID: newOrder.Id,
		CreatedAt: newOrder.CreatedAt.AsTime(),
		AccountID: newOrder.AccountId,
		TotalPrice: newOrder.TotalPrice,
		Products: products,
	}, nil


}

func(c *Client) GetOrdersForAccount(ctx context.Context, id string) ([]Order, error) {
	grpcReq := &pb.GetOrdersForAccountRequest{
		AccountId: id,
	}

	resp, err := c.service.GetOrdersForAccount(ctx, grpcReq)
	if err != nil {
		log.Printf("GetOrdersForAccount RPC failed: %v\n", err)		
		return nil, err
	}

	orders := []Order{}
	for _, o := range resp.Orders {
		order := Order{
			ID: o.Id,
			AccountID: o.AccountId,
			TotalPrice: o.TotalPrice,
			CreatedAt: o.CreatedAt.AsTime(),
			Products: []OrderedProduct{},
		}
		for _, op := range o.Products {
			ordered := OrderedProduct{
				ProductID: op.ProductId,
				Name: op.Name,
				Description: op.Description,
				Price: op.Price,
				Quantity: op.Quantity,
			}
			order.Products = append(order.Products, ordered)
		}
		orders = append(orders, order)
	}
	return orders, nil
} 
