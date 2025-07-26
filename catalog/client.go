package catalog

import (
	"context"

	"github.com/zenvisjr/building-scalable-microservices/catalog/pb"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.CatalogServiceClient
	logs    *logger.Logs
}

func NewClient(address string) (*Client, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Connecting to Catalog gRPC service at " + address)

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		Logs.Error(context.Background(), "Failed to connect to catalog gRPC: "+err.Error())
		return nil, err
	}

	Logs.LocalOnlyInfo("Connected to Catalog gRPC service")

	service := pb.NewCatalogServiceClient(conn)
	return &Client{
		conn:    conn,
		service: service,
		logs:    Logs,
	}, nil
}

func (c *Client) Close() {
	c.logs.LocalOnlyInfo("Closing Catalog client connection")
	c.conn.Close()
}

func (c *Client) PostProduct(ctx context.Context, name, description string, price float64) (*Product, error) {
	c.logs.Info(ctx, "Posting new product to catalog")

	req := &pb.PostProductRequest{
		Name:        name,
		Description: description,
		Price:       price,
	}

	resp, err := c.service.PostProduct(ctx, req)
	if err != nil {
		c.logs.Error(ctx, "PostProduct failed: "+err.Error())
		return nil, err
	}

	c.logs.Info(ctx, "Product posted successfully with ID: "+resp.Product.Id)

	return &Product{
		ID:          resp.Product.Id,
		Name:        resp.Product.Name,
		Description: resp.Product.Description,
		Price:       resp.Product.Price,
	}, nil
}

func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	c.logs.Info(ctx, "Fetching product by ID: "+id)

	req := &pb.GetProductRequest{Id: id}
	resp, err := c.service.GetProduct(ctx, req)
	if err != nil {
		c.logs.Error(ctx, "GetProduct failed: "+err.Error())
		return nil, err
	}

	c.logs.Info(ctx, "Product fetched: " + resp.Product.Name)

	return &Product{
		ID:          resp.Product.Id,
		Name:        resp.Product.Name,
		Description: resp.Product.Description,
		Price:       resp.Product.Price,
	}, nil
}

func (c *Client) GetProducts(ctx context.Context, skip uint64, take uint64, ids []string, query string) ([]Product, error) {
	c.logs.Info(ctx, "Fetching products with pagination")

	req := &pb.GetProductsRequest{
		Skip:  skip,
		Take:  take,
		Ids:   ids,
		Query: query,
	}

	resp, err := c.service.GetProducts(ctx, req)
	if err != nil {
		c.logs.Error(ctx, "GetProducts failed: "+err.Error())
		return nil, err
	}

	c.logs.Info(ctx, "Fetched products from catalog: count = "+logger.IntToStr(len(resp.Products)))

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
