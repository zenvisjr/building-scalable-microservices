package catalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/olivere/elastic/v7"
	"github.com/zenvisjr/building-scalable-microservices/logger"
)

var (
	errNotFound   = fmt.Errorf("entity not found")
	errOutOfStock = fmt.Errorf("product is out of stock")
)

type Repository interface {
	// Close()
	PutProduct(ctx context.Context, product Product) error
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
	UpdateStockAndSold(ctx context.Context, id string, quantity int) (bool, error)
}

type elasticRepository struct {
	client *elastic.Client
}

type productDocument struct {
	Name        string  `json:"name"`         // Product name
	Description string  `json:"description"`  // Product description
	Price       float64 `json:"price"`        // Product price
	Stock       uint32  `json:"stock"`        // Available stock for order
	Sold        uint32  `json:"sold"`         // Total units sold
	OutOfStock  bool    `json:"out_of_stock"` // Product is out of stock
}

func NewElasticRepository(url string) (Repository, error) {
	ctx := context.Background()
	//Default Connect to localhost:9200
	//here we provide the url
	//“Connect to this Elasticsearch host (on port 9200 or whatever the URL says), but don’t try to sniff and discover other nodes.”
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Connecting to Elasticsearch at "+url)
	client, err := elastic.NewClient(elastic.SetURL(url), elastic.SetSniff(false))
	if err != nil {
		Logs.Error(ctx, "Failed to connect to Elasticsearch: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Connected to Elasticsearch at "+url)
	Logs.RemoteLogs(ctx, "Connected to Elasticsearch at "+url, "catalog")
	return &elasticRepository{client}, nil
}

// func (p *elasticRepository) Close() {
// 	p.client.
// }

func (p *elasticRepository) PutProduct(ctx context.Context, product Product) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Indexing product: " + product.ID)

	document := productDocument{
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		Sold:        0,
		OutOfStock:  false,
	}
	res, err := p.client.Index().Index("catalog").Id(product.ID).BodyJson(document).Do(ctx)
	if err != nil {
		Logs.Error(ctx, "Failed to index product: "+err.Error())
		return err
	}

	Logs.LocalOnlyInfo("Product indexed successfully: " + product.ID)
	Logs.RemoteLogs(ctx, "Product indexed: "+product.ID, "catalog")
	fmt.Println(res)
	return nil
}

func (p *elasticRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Fetching product by ID: "+id)

	//fetch the product from elasticsearch
	res, err := p.client.Get().Index("catalog").Id(id).Do(ctx)
	if err != nil {
		Logs.Error(ctx, "Failed to fetch product by ID: "+err.Error())
		return nil, err
	}
	if !res.Found {
		Logs.Error(ctx, "Product not found: "+id)
		return nil, errNotFound
	}
	var doc productDocument
	if err := json.Unmarshal(res.Source, &doc); err != nil {
		Logs.Error(ctx, "Failed to unmarshal product: "+err.Error())
		return nil, err
	}

	//checking if product is out of stock
	if doc.OutOfStock {
		Logs.Error(ctx, "Product is out of stock: "+id)
		return nil, errOutOfStock
	}
	Logs.Info(ctx, "Product fetched: "+id)
	return &Product{
		ID:          id,
		Name:        doc.Name,
		Description: doc.Description,
		Price:       doc.Price,
		Stock:       doc.Stock,
		Sold:        doc.Sold,
	}, nil
}

func (p *elasticRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Listing products")
	res, err := p.client.Search().Index("catalog").Query(elastic.NewMatchAllQuery()).Size(int(take)).From(int(skip)).Do(ctx)
	if err != nil {
		Logs.Error(ctx, "Failed to list products: "+err.Error())
		return nil, err
	}
	products := []Product{}
	for _, hit := range res.Hits.Hits {
		product := &productDocument{}
		if err := json.Unmarshal(hit.Source, &product); err != nil {
			continue
		}
		//checking if product is out of stock
		if product.OutOfStock {
			Logs.Error(ctx, "Product is out of stock: "+hit.Id)
			continue
		}

		products = append(products, Product{
			Name:        product.Name,
			ID:          hit.Id,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			Sold:        product.Sold,
		})
	}
	Logs.Info(ctx, "Products listed: "+logger.IntToStr(len(products)))
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

//		return products, nil
//	}
func (r *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Listing products with IDs: " + logger.IntToStr(len(ids)))

	items := []*elastic.MultiGetItem{}
	for _, id := range ids {
		items = append(items, elastic.NewMultiGetItem().Index("catalog").Id(id))
	}

	res, err := r.client.MultiGet().Add(items...).Do(ctx)
	if err != nil {
		Logs.Error(ctx, "MultiGet error: "+err.Error())
		return nil, err
	}

	Logs.LocalOnlyInfo("MultiGet returned " + logger.IntToStr(len(res.Docs)) + " docs")

	products := []Product{}
	for _, doc := range res.Docs {
		if !doc.Found || doc.Source == nil {
			Logs.Error(ctx, "Missing doc or nil source for ID: "+doc.Id)
			continue
		}
		p := productDocument{}
		if err := json.Unmarshal(doc.Source, &p); err != nil {
			Logs.Error(ctx, "Unmarshal failed for ID: "+doc.Id+" | "+err.Error())
			continue
		}
		//checking if product is out of stock
		if p.OutOfStock {
			Logs.Error(ctx, "Product is out of stock: "+doc.Id)
			continue
		}
		products = append(products, Product{
			ID:          doc.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			Sold:        p.Sold,
		})
	}

	Logs.LocalOnlyInfo("Fetched " + logger.IntToStr(len(products)) + " products by ID")
	return products, nil
}

func (p *elasticRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	//we are seraching product accross multiple fields by matching it against name
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Searching products | query: \""+query+"\", skip: "+logger.Uint64ToStr(skip)+", take: "+logger.Uint64ToStr(take))
	Logs.RemoteLogs(ctx, "Searching products | query: \""+query+"\", skip: "+logger.Uint64ToStr(skip)+", take: "+logger.Uint64ToStr(take), "catalog")
	res, err := p.client.Search().Index("catalog").Query(elastic.NewMultiMatchQuery(query, "name", "description")).Size(int(take)).From(int(skip)).Do(ctx)
	if err != nil {
		Logs.Error(ctx, "Failed to search products: "+err.Error())
		Logs.RemoteLogs(ctx, "Failed to search products: "+err.Error(), "catalog")
		return nil, err
	}
	products := []Product{}
	for _, hit := range res.Hits.Hits {
		product := &productDocument{}
		if err := json.Unmarshal(hit.Source, &product); err != nil {
			continue
		}
		//checking if product is out of stock
		if product.OutOfStock {
			Logs.Error(ctx, "Product is out of stock: "+hit.Id)
			continue
		}
		products = append(products, Product{
			Name:        product.Name,
			ID:          hit.Id,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			Sold:        product.Sold,
		})
	}

	Logs.Info(ctx, "Products searched: "+logger.IntToStr(len(products)))
	Logs.RemoteLogs(ctx, "Products searched: "+logger.IntToStr(len(products)), "catalog")
	return products, nil
}

func (p *elasticRepository) UpdateStockAndSold(ctx context.Context, id string, quantity int) (bool, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Updating stock and sold for product: "+id)

	// Step 1: Define the script
	script := elastic.NewScriptInline(`
		if (ctx._source.stock < params.qty) {
			throw new Exception("insufficient stock");
		}
		ctx._source.stock -= params.qty;
		ctx._source.sold += params.qty;
		if (ctx._source.stock <= 0) {
			ctx._source.outOfStock = true;
		}
	`).Lang("painless").Params(map[string]interface{}{
		"qty": quantity,
	})

	// Step 2: Execute the update script
	_, err := p.client.Update().
		Index("catalog").
		Id(id).
		Script(script).
		Do(ctx)

	if err != nil {
		Logs.Error(ctx, "Failed to update stock and sold using script: "+err.Error())
		return false, err
	}

	Logs.Info(ctx, "Successfully updated stock and sold for product: "+id)
	Logs.RemoteLogs(ctx, "Stock and sold updated for product: "+id, "catalog")
	return true, nil
}

// What is sniffing?
// By default, the client tries to discover all nodes in your cluster by calling:
// GET /_nodes/http
//is recommended in production
