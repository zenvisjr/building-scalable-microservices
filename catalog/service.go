package catalog

import (
	"context"

	"github.com/segmentio/ksuid"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    uint32  `json:"quantity"`
}

type Service interface {
	PostProduct(ctx context.Context, name, description string, price float64) (*Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	GetProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	GetProductsByIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type catalogService struct {
	repo Repository
}

func NewCatalogService(repo Repository) Service {
	return &catalogService{repo: repo}
}

func (s *catalogService) PostProduct(ctx context.Context, name, description string, price float64) (*Product, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Creating new product")

	product := Product{
		ID:          ksuid.New().String(),
		Name:        name,
		Description: description,
		Price:       price,
	}
	if err := s.repo.PutProduct(ctx, product); err != nil {
		Logs.Error(ctx, "Failed to store new product: "+err.Error())
		return nil, err
	}

	Logs.LocalOnlyInfo("Product created with ID: " + product.ID)
	return &product, nil
}

func (s *catalogService) GetProduct(ctx context.Context, id string) (*Product, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Retrieving product by ID: " + id)

	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		Logs.Error(ctx, "Failed to retrieve product ID "+id+": "+err.Error())
		return nil, err
	}
	return product, nil
}

func (s *catalogService) GetProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	Logs := logger.GetGlobalLogger()
	if take > 100 || (take == 0 && skip == 0) {
		take = 100
	}
	Logs.LocalOnlyInfo("Listing products | skip: " + logger.Uint64ToStr(skip) + ", take: " + logger.Uint64ToStr(take))

	products, err := s.repo.ListProducts(ctx, skip, take)
	if err != nil {
		Logs.Error(ctx, "Failed to list products: "+err.Error())
	}
	return products, err
}

func (s *catalogService) GetProductsByIDs(ctx context.Context, ids []string) ([]Product, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Fetching products by multiple IDs (count: " + logger.IntToStr(len(ids)) + ")")

	products, err := s.repo.ListProductsWithIDs(ctx, ids)
	if err != nil {
		Logs.Error(ctx, "Failed to fetch products by IDs: "+err.Error())
	}
	return products, err
}

func (s *catalogService) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	Logs := logger.GetGlobalLogger()
	if take > 100 || (take == 0 && skip == 0) {
		take = 100
	}
	Logs.LocalOnlyInfo("Searching products | query: \"" + query + "\", skip: " + logger.Uint64ToStr(skip) + ", take: " + logger.Uint64ToStr(take))

	products, err := s.repo.SearchProducts(ctx, query, skip, take)
	if err != nil {
		Logs.Error(ctx, "Search failed: "+err.Error())
	}
	return products, err
}
