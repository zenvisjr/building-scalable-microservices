package order

import (
	"context"

	"github.com/zenvisjr/building-scalable-microservices/logger"
	"github.com/zenvisjr/building-scalable-microservices/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(address string) (*Client, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Connecting to OrderService at " + address)

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		Logs.Error(context.Background(), "Failed to create gRPC connection: "+err.Error())
		return nil, err
	}

	Logs.Info(context.Background(), "Connected to OrderService at " + address)
	return &Client{
		conn:    conn,
		service: pb.NewOrderServiceClient(conn),
	}, nil
}

func (c *Client) Close() {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Closing gRPC connection to OrderService")
	c.conn.Close()
}

func (c *Client) PostOrder(ctx context.Context, id string, products []OrderedProduct) (*Order, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Entered PostOrder()")

	Logs.LocalOnlyInfo("Preparing gRPC product list from input")
	productList := []*pb.PostOrderRequest_OrderedProduct{}
	for _, p := range products {
		Logs.LocalOnlyInfo("Processing product: " + p.ProductID)
		productList = append(productList, &pb.PostOrderRequest_OrderedProduct{
			ProductId: p.ProductID,
			Quantity:  p.Quantity,
		})
	}

	Logs.LocalOnlyInfo("Creating gRPC PostOrderRequest")
	grpcReq := &pb.PostOrderRequest{
		AccountId: id,
		Products:  productList,
	}

	Logs.LocalOnlyInfo("Sending PostOrder RPC")
	resp, err := c.service.PostOrder(ctx, grpcReq)
	if err != nil {
		Logs.Error(ctx, "PostOrder RPC failed: "+err.Error())
		return nil, err
	}
	Logs.LocalOnlyInfo("PostOrder RPC successful")

	newOrder := resp.Order
	Logs.LocalOnlyInfo("Parsing gRPC response to internal struct")

	ordered := []OrderedProduct{}
	for _, p := range newOrder.Products {
		Logs.LocalOnlyInfo("Parsing product from gRPC: " + p.ProductId)
		ordered = append(ordered, OrderedProduct{
			ProductID:   p.ProductId,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
			Stock:       p.Stock,
		})
	}

	Logs.LocalOnlyInfo("Order struct fully built")
	Logs.Info(ctx, "Order created with ID: "+newOrder.Id+" for AccountID: "+id)

	return &Order{
		ID:         newOrder.Id,
		CreatedAt:  newOrder.CreatedAt.AsTime(),
		AccountID:  newOrder.AccountId,
		TotalPrice: newOrder.TotalPrice,
		Products:   ordered,
	}, nil
}

func (c *Client) GetOrdersForAccount(ctx context.Context, id string) ([]Order, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Entered GetOrdersForAccount() for ID: " + id)

	grpcReq := &pb.GetOrdersForAccountRequest{
		AccountId: id,
	}

	Logs.LocalOnlyInfo("Sending GetOrdersForAccount RPC")
	resp, err := c.service.GetOrdersForAccount(ctx, grpcReq)
	if err != nil {
		Logs.Error(ctx, "GetOrdersForAccount RPC failed: "+err.Error())
		return nil, err
	}
	Logs.LocalOnlyInfo("GetOrdersForAccount RPC successful")

	Logs.Info(ctx, "Fetched "+logger.IntToStr(len(resp.Orders))+" orders for AccountID: "+id)

	orders := []Order{}
	for _, o := range resp.Orders {
		order := Order{
			ID:         o.Id,
			AccountID:  o.AccountId,
			TotalPrice: o.TotalPrice,
			CreatedAt:  o.CreatedAt.AsTime(),
			Products:   []OrderedProduct{},
		}
		for _, op := range o.Products {
			order.Products = append(order.Products, OrderedProduct{
				ProductID:   op.ProductId,
				Name:        op.Name,
				Description: op.Description,
				Price:       op.Price,
				Quantity:    op.Quantity,
				Stock:       op.Stock,
			})
		}
		orders = append(orders, order)
	}

	Logs.LocalOnlyInfo("Successfully parsed all orders")
	Logs.Info(ctx, "Fetched "+logger.IntToStr(len(resp.Orders))+" orders for AccountID: "+id)
	return orders, nil
}
