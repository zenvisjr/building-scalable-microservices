package catalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"
)

var (
	errNotFound = fmt.Errorf("Entity not found")
)

type Repository interface {
	// Close()
	PutProduct(ctx context.Context, product Product) error
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type elasticRepository struct {
	client *elastic.Client
}

type productDocument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       float64 `json:"price"`
}

func NewElasticRepository(url string) (Repository, error) {
	//Default Connect to localhost:9200
	//here we provide the url
	//“Connect to this Elasticsearch host (on port 9200 or whatever the URL says), but don’t try to sniff and discover other nodes.”
	client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(false))
	if err != nil {
		return nil, err
	}

	return &elasticRepository{client}, nil
}

// func (p *elasticRepository) Close() {
// 	p.client.
// }

func (p *elasticRepository) PutProduct(ctx context.Context, product Product) error {
	document := productDocument{
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	}
	res, err := p.client.Index().Index("catalog").Id(product.ID).BodyJson(document).Do(ctx)
	if err != nil {
		return err
	}
	fmt.Println(res)
	return nil
}

func (p *elasticRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	res, err := p.client.Get().Index("catalog").Id(id).Do(ctx)
	if err != nil {
		return nil, err
	}
	if !res.Found {
		return nil, errNotFound
	}
	var doc productDocument
	if err := json.Unmarshal(res.Source, &doc); err != nil {
		return nil, err
	}

	return &Product{
		ID:          id,
		Name:        doc.Name,
		Description: doc.Description,
		Price:       doc.Price,
	}, nil
}

func (p *elasticRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	res, err := p.client.Search().Index("catalog").Query(elastic.NewMatchAllQuery()).Size(int(take)).From(int(skip)).Do(ctx)
	if err != nil {
		return nil, err
	}
	products := []Product{}
	for _, hit := range res.Hits.Hits {
		product := &productDocument{}
		if err := json.Unmarshal(hit.Source, &product); err != nil {
			continue
		}
		products = append(products, Product{
			Name:        product.Name,
			ID:          hit.Id,
			Description: product.Description,
			Price:       product.Price,
		})
	}

	return products, nil
}

func (p *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {

	term := make([]interface{}, len(ids))
	for i, id := range ids {
		term[i] = id
	}
	res, err := p.client.Search().Index("catalog").Query(elastic.NewTermsQuery("id.keyword", term...)).Do(ctx)
	if err != nil {
		return nil, err
	}
	products := []Product{}
	for _, hit := range res.Hits.Hits {
		product := &productDocument{}
		if err := json.Unmarshal(hit.Source, &product); err != nil {
			continue
		}
		products = append(products, Product{
			Name:        product.Name,
			ID:          hit.Id,
			Description: product.Description,
			Price:       product.Price,
		})
	}

	return products, nil
}

func (p *elasticRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	//we are seraching product accross multiple fields by matching it against name
	res, err := p.client.Search().Index("catalog").Query(elastic.NewMultiMatchQuery(query, "name", "description")).Size(int(take)).From(int(skip)).Do(ctx)
	if err != nil {
		return nil, err
	}
	products := []Product{}
	for _, hit := range res.Hits.Hits {
		product := &productDocument{}
		if err := json.Unmarshal(hit.Source, &product); err != nil {
			continue
		}
		products = append(products, Product{
			Name:        product.Name,
			ID:          hit.Id,
			Description: product.Description,
			Price:       product.Price,
		})
	}

	return products, nil
}

// What is sniffing?
// By default, the client tries to discover all nodes in your cluster by calling:
// GET /_nodes/http
//is recommended in production
