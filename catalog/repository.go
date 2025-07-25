package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
	Name        string  `json:"name"`
	Description string  `json:"description"`
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

// func (p *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
// 	log.Printf("🔥 finally ListProductsWithIDs called with Ids=%v", ids)
// 	term := make([]interface{}, len(ids))
// 	for i, id := range ids {
// 		term[i] = id
// 	}
// 	log.Printf("🔥 finally ListProductsWithIDs called with Ids=%v after creating term", term)
// 	res, err := p.client.Search().Index("catalog").Query(elastic.NewTermsQuery("id.keyword", term...)).Do(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Printf("we get response from elasticsearch %v with length %d", res, len(res.Hits.Hits))
// 	products := []Product{}
// 	for _, hit := range res.Hits.Hits {
// 		product := &productDocument{}
// 		if err := json.Unmarshal(hit.Source, &product); err != nil {
// 			continue
// 		}
// 		products = append(products, Product{
// 			Name:        product.Name,
// 			ID:          hit.Id,
// 			Description: product.Description,
// 			Price:       product.Price,
// 		})
// 	}
// 	log.Printf("no of products fetched from elasticsearch %v", len(products))

// 	return products, nil
// }

func (r *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	log.Printf("🔥 ListProductsWithIDs called with Ids=%v", ids)
	
	items := []*elastic.MultiGetItem{}
	for _, id := range ids {
		items = append(
			items,
			elastic.NewMultiGetItem().
				Index("catalog").
				// Type("product").
				Id(id),
		)
	}
	
	res, err := r.client.MultiGet().Add(items...).Do(ctx)
	if err != nil {
		log.Printf("❌ MultiGet error: %v", err)
		return nil, err
	}
	
	log.Printf("🔥 MultiGet returned %d docs", len(res.Docs))
	
	products := []Product{}
	for i, doc := range res.Docs {
		log.Printf("🔥 Processing doc %d:", i)
		log.Printf("🔥   Doc ID: %s", doc.Id)
		log.Printf("🔥   Doc Found: %t", doc.Found)
		log.Printf("🔥   Doc Source: %s", string(doc.Source))
		
		if !doc.Found {
			log.Printf("❌ Document not found for ID: %s", doc.Id)
			continue
		}
		
		if doc.Source == nil {
			log.Printf("❌ Document source is nil for ID: %s", doc.Id)
			continue
		}
		
		p := productDocument{}
		if err = json.Unmarshal(doc.Source, &p); err != nil {
			log.Printf("❌ JSON unmarshal error: %v", err)
			log.Printf("❌ Raw source: %s", string(doc.Source))
			continue
		}
		
		log.Printf("✅ Successfully unmarshaled: Name=%s, Price=%f", p.Name, p.Price)
		
		products = append(products, Product{
			ID:          doc.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		})
	}
	
	log.Printf("🔥 Final result: %d products", len(products))
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
