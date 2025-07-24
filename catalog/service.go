package catalog

import (
	"context"

	"github.com/segmentio/ksuid"
)

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       float64 `json:"price"`
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
	product := Product {
		ID: ksuid.New().String(),
		Name: name,
		Description: description,
		Price: price,
	}
	if err := s.repo.PutProduct(ctx, product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *catalogService) GetProduct(ctx context.Context, id string) (*Product, error) {
	return s.repo.GetProductByID(ctx, id)
}

func (s *catalogService) GetProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (take == 0 && skip == 0) {
		take = 100
	}
		
	
	return s.repo.ListProducts(ctx, skip, take	)
}

func (s *catalogService) GetProductsByIDs(ctx context.Context, ids []string) ([]Product, error) {
	return s.repo.ListProductsWithIDs(ctx, ids)
}

func (s *catalogService) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (take == 0 && skip == 0) {
		take = 100
	}
	return s.repo.SearchProducts(ctx, query, skip, take)
}
