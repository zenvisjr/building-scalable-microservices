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

func (c *Client) PostProduct(ctx context.Context, name, description string, price float64, stock int) (*Product, error) {
	c.logs.Info(ctx, "Posting new product to catalog")

	req := &pb.PostProductRequest{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       uint32(stock),
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
		Stock:       uint32(resp.Product.Stock),
		Sold:        uint32(resp.Product.Sold),
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
		Stock:       uint32(resp.Product.Stock),
		Sold:        uint32(resp.Product.Sold),
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
			Stock:       uint32(p.Stock),
			Sold:        uint32(p.Sold),
			OutOfStock:  p.OutOfStock,
		}
	}

	return products, nil
}

func (c *Client) UpdateStockAndSold(ctx context.Context, id string, quantity int) (bool, error) {
	c.logs.Info(ctx, "Updating stock and sold for product: "+id)

	req := &pb.UpdateStockRequest{
		ProductId: id,
		Quantity:  int32(quantity),
	}

	resp, err := c.service.UpdateStockAndSold(ctx, req)
	if err != nil {
		c.logs.Error(ctx, "UpdateStockAndSold failed: "+err.Error())
		return false, err
	}

	c.logs.Info(ctx, "Stock and sold updated for product: "+id)

	return resp.Ok, nil
}

func (c *Client) DeleteProduct(ctx context.Context, id string) error {
	c.logs.Info(ctx, "Soft-deleting product (set outOfStock=true): "+id)

	req := &pb.DeleteProductRequest{Id: id}
	_, err := c.service.DeleteProduct(ctx, req)
	if err != nil {
		c.logs.Error(ctx, "DeleteProduct failed: "+err.Error())
		return err
	}

	c.logs.Info(ctx, "Product soft-deleted (outOfStock=true): "+id)
	return nil
}

func (c *Client) RestockProduct(ctx context.Context, id string, newStock int) error {
	c.logs.Info(ctx, "Restocking product: "+id)

	req := &pb.RestockProductRequest{
		ProductId: id,
		NewStock:  int32(newStock),
	}

	_, err := c.service.RestockProduct(ctx, req)
	if err != nil {
		c.logs.Error(ctx, "RestockProduct failed: "+err.Error())
		return err
	}

	c.logs.Info(ctx, "Product restocked: "+id)
	return nil
}

func (c *Client) SuggestProducts(ctx context.Context, prefix string, size int, useAI bool) ([]Product, error) {
	c.logs.Info(ctx, "Suggesting products with prefix: "+prefix)

	req := &pb.SuggestProductsRequest{
		Query: prefix,
		Size:  int32(size),
		UseAi: useAI,
	}

	resp, err := c.service.SuggestProducts(ctx, req)
	if err != nil {
		c.logs.Error(ctx, "SuggestProducts failed: "+err.Error())
		return nil, err
	}

	c.logs.Info(ctx, "Products suggested: count = "+logger.IntToStr(len(resp.Products)))

	products := make([]Product, len(resp.Products))
	for i, p := range resp.Products {
		products[i] = Product{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       uint32(p.Stock),
			Sold:        uint32(p.Sold),
			OutOfStock:  p.OutOfStock,
		}
	}

	return products, nil
}
