package catalog

import (
	"context"

	"github.com/zenvisjr/building-scalable-microservices/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.CatalogServiceClient
}

func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	service := pb.NewCatalogServiceClient(conn)
	return &Client{
		conn:    conn,
		service: service,
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostProduct(ctx context.Context, name, description string, price float64) (*Product, error) {
	req := &pb.PostProductRequest{
		Name:        name,
		Description: description,
		Price:       price,
	}

	resp, err := c.service.PostProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          resp.Product.Id,
		Name:        resp.Product.Name,
		Description: resp.Product.Description,
		Price:       resp.Product.Price,
	}, nil
}

func(c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	req := &pb.GetProductRequest{
		Id: id,
	}

	resp, err := c.service.GetProduct(ctx, req)
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          resp.Product.Id,
		Name:        resp.Product.Name,
		Description: resp.Product.Description,
		Price:       resp.Product.Price,
	}, nil
}

func(c *Client) GetProducts(ctx context.Context, skip uint64, take uint64, ids []string, query string) ([]Product, error) {

	req := &pb.GetProductsRequest{
			Skip: skip,
			Take: take,
			Ids: ids,
			Query: query,
	}

	resp, err := c.service.GetProducts(ctx, req)
	if err != nil {
		return nil, err
	}
	products := make([]Product, len(resp.Products))
	for i, p := range resp.Products {
		products[i] = Product{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}
	}
	return products, nil
}

