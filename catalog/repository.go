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
	CreateProduct(ctx context.Context, product Product) error
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
	UpdateStockAndSold(ctx context.Context, id string, quantity int) (bool, error)
	DeleteProductByID(ctx context.Context, id string) error
	RestockProduct(ctx context.Context, id string, newStock int) error
	EnsureCatalogIndex(ctx context.Context) error
	CreateCatalogIndexWithAutocomplete(ctx context.Context) error
	SuggestProducts(ctx context.Context, prefix string, size int) ([]Product, error)
	AISuggest(ctx context.Context, query string, size int) ([]Product, error)
}

type elasticRepository struct {
	client *elastic.Client
}

type productDocument struct {
	Name        string    `json:"name"`         // Product name
	Description string    `json:"description"`  // Product description
	Price       float64   `json:"price"`        // Product price
	Stock       uint32    `json:"stock"`        // Available stock for order
	Sold        uint32    `json:"sold"`         // Total units sold
	OutOfStock  bool      `json:"out_of_stock"` // Product is out of stock
	Embedding   []float64 `json:"embedding"`    // OpenAI embedding vector
}

func NewElasticRepository(url string) (Repository, error) {
	ctx := context.Background()
	//Default Connect to localhost:9200
	//here we provide the url
	//“Connect to this Elasticsearch host (on port 9200 or whatever the URL says), but don’t try to sniff and discover other nodes.”
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Connecting to Elasticsearch at "+url)
	client, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetSniff(false),
		elastic.SetBasicAuth("elastic", "123456"),
	)
	if err != nil {
		Logs.Error(ctx, "Failed to connect to Elasticsearch: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Connected to Elasticsearch at "+url)
	return &elasticRepository{client}, nil
}

// func (p *elasticRepository) Close() {
// 	p.client.Close()
// }

// func (p *elasticRepository) PutProduct(ctx context.Context, product Product) error {
// 	Logs := logger.GetGlobalLogger()
// 	// Logs.Info(ctx, "Getting embedding for product: "+product.ID)
// 	// embedding, err := GetEmbeddingFromPython(product.Name, product.Description)
// 	// if err != nil {
// 	// 	Logs.Error(ctx, "Failed to get embedding: "+err.Error())
// 	// 	return err
// 	// }
// 	Logs.LocalOnlyInfo("Indexing product: " + product.ID)

// 	document := productDocument{
// 		Name:        product.Name,
// 		Description: product.Description,
// 		Price:       product.Price,
// 		Stock:       product.Stock,
// 		Sold:        0,
// 		OutOfStock:  false,
// 		// Embedding:   embedding,
// 	}

// 	indexCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
// 	defer cancel()

// 	response, err := p.client.Index().
// 		Index("catalog").
// 		Id(product.ID).
// 		BodyJson(document).
// 		Do(indexCtx)

// 	if err != nil {
// 		Logs.Error(ctx, "Failed to index product: "+err.Error())
// 		return err
// 	}

// 	// Check if the response indicates success
// 	if response == nil {
// 		Logs.Error(ctx, "Elasticsearch returned nil response for product: "+product.ID)
// 		return fmt.Errorf("elasticsearch returned nil response")
// 	}

// 	Logs.LocalOnlyInfo("Product indexed successfully: " + product.ID)

// 	go p.AddEmbeddingToProduct(product)

// 	// Logs.Info(ctx, res)
// 	return nil
// }

// func (p *elasticRepository) PutProduct(ctx context.Context, product Product) error {
// 	Logs := logger.GetGlobalLogger()
// 	Logs.LocalOnlyInfo("Indexing product: " + product.ID)

// 	// Check if client exists
// 	if p.client == nil {
// 		Logs.LocalOnlyInfo("ERROR: Elasticsearch client is nil!")
// 		return fmt.Errorf("elasticsearch client is nil")
// 	}
// 	Logs.LocalOnlyInfo("Client exists, proceeding...")

// 	document := productDocument{
// 		Name:        product.Name,
// 		Description: product.Description,
// 		Price:       product.Price,
// 		Stock:       product.Stock,
// 		Sold:        0,
// 		OutOfStock:  false,
// 	}
// 	Logs.LocalOnlyInfo("Document created: " + fmt.Sprintf("%+v", document))

// 	// Test basic connectivity first
// 	// Logs.LocalOnlyInfo("Testing Elasticsearch connectivity...")
// 	// pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	// defer pingCancel()

// 	// _, code, pingErr := p.client.Ping("http://catalog_db:9200").Do(pingCtx)
// 	// if pingErr != nil {
// 	// 	Logs.LocalOnlyInfo("PING FAILED: " + pingErr.Error())
// 	// 	return fmt.Errorf("elasticsearch ping failed: %v", pingErr)
// 	// }
// 	// Logs.LocalOnlyInfo("Ping successful, status code: " + fmt.Sprintf("%d", code))

// 	// Create index request
// 	Logs.LocalOnlyInfo("Creating index request...")
// 	indexReq := p.client.Index().Index("catalog").Id(product.ID).BodyJson(document)
// 	Logs.LocalOnlyInfo("Index request created")

// 	// Add timeout context
// 	// indexCtx, cancel := context.WithTimeout(context.Background(), 0*time.Second)
// 	// defer cancel()

// 	Logs.LocalOnlyInfo("About to call Do() on index request...")

// 	// This is where it's likely hanging - let's see
// 	response, err := indexReq.Do(ctx)

// 	Logs.LocalOnlyInfo("Do() call completed!")

// 	if err != nil {
// 		Logs.LocalOnlyInfo("ERROR in Do(): " + err.Error())
// 		return err
// 	}

// 	if response == nil {
// 		Logs.LocalOnlyInfo("ERROR: Response is nil")
// 		return fmt.Errorf("elasticsearch returned nil response")
// 	}

// 	Logs.LocalOnlyInfo("SUCCESS: Product indexed, Result: " + response.Result)
// 	return nil
// }

func (p *elasticRepository) CreateProduct(ctx context.Context, product Product) error {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Received product for indexing: "+product.ID)

	document := productDocument{
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		Sold:        0,
		OutOfStock:  false,
	}
	Logs.Info(ctx, "Indexing product: "+product.ID)
	response, err := p.client.Index().
		Index("catalog").
		Id(product.ID).
		BodyJson(document).
		Do(ctx)

	if err != nil {
		Logs.Error(ctx, "Failed to index product: "+product.ID+": "+err.Error())
		return err
	}

	if response != nil {
		Logs.Info(ctx, "Product indexed successfully: "+product.ID)
	}
	Logs.Info(ctx, "Product queued for async embedding: "+product.ID)
	go p.AddEmbeddingToProduct(product)

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
			OutOfStock:  product.OutOfStock,
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
	res, err := p.client.Search().Index("catalog").Query(elastic.NewMultiMatchQuery(query, "name", "description")).Size(int(take)).From(int(skip)).Do(ctx)
	if err != nil {
		Logs.Error(ctx, "Failed to search products: "+err.Error())
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
	return true, nil
}

func (p *elasticRepository) DeleteProductByID(ctx context.Context, id string) error {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Soft-deleting product (set outOfStock=true): "+id)

	// Step 1: Check if product exists
	exists, err := p.client.Exists().
		Index("catalog").
		Id(id).
		Do(ctx)
	if err != nil {
		Logs.Error(ctx, "Error checking product existence: "+err.Error())
		return err
	}
	if !exists {
		Logs.Error(ctx, "Product not found for soft delete: "+id)
		return errNotFound
	}

	// Step 2: Update outOfStock = true and stock = 0
	script := elastic.NewScriptInline(`
		ctx._source.out_of_stock = true;
		ctx._source.stock = 0;
	`).Lang("painless")

	_, err = p.client.Update().
		Index("catalog").
		Id(id).
		Script(script).
		Refresh("true").
		Do(ctx)

	if err != nil {
		Logs.Error(ctx, "Failed to soft-delete product: "+err.Error())
		return err
	}

	Logs.Info(ctx, "Product soft-deleted (outOfStock=true): "+id)
	return nil
}

func (p *elasticRepository) RestockProduct(ctx context.Context, id string, newStock int) error {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Restocking product: "+id)

	// Check if product exists
	exists, err := p.client.Exists().Index("catalog").Id(id).Do(ctx)
	if err != nil {
		Logs.Error(ctx, "Error checking product existence: "+err.Error())
		return err
	}
	if !exists {
		Logs.Error(ctx, "Product not found for restock: "+id)
		return errNotFound
	}

	// Prepare script to restock
	script := elastic.NewScriptInline(`
		ctx._source.stock = params.newStock;
		ctx._source.out_of_stock = false;
	`).Lang("painless").Params(map[string]interface{}{
		"newStock": newStock,
	})

	// Perform update
	_, err = p.client.Update().
		Index("catalog").
		Id(id).
		Script(script).
		Refresh("true").
		Do(ctx)
	if err != nil {
		Logs.Error(ctx, "Failed to restock product: "+err.Error())
		return err
	}

	Logs.Info(ctx, "Successfully restocked product: "+id)
	Logs.RemoteLogs(ctx, "Product restocked: "+id, "catalog")
	return nil
}

// What is sniffing?
// By default, the client tries to discover all nodes in your cluster by calling:
// GET /_nodes/http
//is recommended in production

// EnsureCatalogIndex checks if the catalog index exists and creates it if it doesn't
func (p *elasticRepository) EnsureCatalogIndex(ctx context.Context) error {
	Logs := logger.GetGlobalLogger()

	// Check if catalog index exists
	exists, err := p.client.IndexExists("catalog").Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if index exists: %w", err)
	}

	if exists {
		Logs.Info(ctx, "Catalog index already exists. Skipping creation.")
		return nil // ✅ Already exists, skip
	}

	Logs.Info(ctx, "Catalog index does not exist. Creating...")
	return p.CreateCatalogIndexWithAutocomplete(ctx)
}

// CreateCatalogIndexWithAutocomplete creates the catalog index with autocomplete analyzer
func (p *elasticRepository) CreateCatalogIndexWithAutocomplete(ctx context.Context) error {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Creating catalog index with autocomplete analyzer")

	// // Delete if exists
	// exists, err := p.client.IndexExists("catalog").Do(ctx)
	// if err != nil {
	// 	return fmt.Errorf("check index exists failed: %w", err)
	// }
	// if exists {
	// 	_, err := p.client.DeleteIndex("catalog").Do(ctx)
	// 	if err != nil {
	// 		return fmt.Errorf("delete old index failed: %w", err)
	// 	}
	// 	Logs.Info(ctx, "Deleted existing catalog index")
	// }

	// Define settings with EdgeNGram analyzer
	createIndex, err := p.client.CreateIndex("catalog").
		BodyJson(map[string]interface{}{
			"settings": map[string]interface{}{
				"analysis": map[string]interface{}{
					"filter": map[string]interface{}{
						"autocomplete_filter": map[string]interface{}{
							"type":     "edge_ngram",
							"min_gram": 2,
							"max_gram": 20,
						},
					},
					"analyzer": map[string]interface{}{
						"autocomplete_analyzer": map[string]interface{}{
							"type":      "custom",
							"tokenizer": "standard",
							"filter":    []string{"lowercase", "autocomplete_filter"},
						},
					},
				},
			},
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":            "text",
						"analyzer":        "autocomplete_analyzer",
						"search_analyzer": "standard",
					},
					"description": map[string]interface{}{
						"type": "text",
					},
					"price": map[string]interface{}{
						"type": "float",
					},
					"stock": map[string]interface{}{
						"type": "integer",
					},
					"sold": map[string]interface{}{
						"type": "integer",
					},
					"out_of_stock": map[string]interface{}{
						"type": "boolean",
					},
					"embedding": map[string]interface{}{
						"type": "dense_vector",
						"dims": 1536,
					},
				},
			},
		}).Do(ctx)

	if err != nil {
		Logs.Error(ctx, "Failed to create catalog index: "+err.Error())
		return err
	}

	if !createIndex.Acknowledged {
		return fmt.Errorf("index creation not acknowledged")
	}

	Logs.Info(ctx, "Catalog index created with autocomplete analyzer")
	return nil
}

func (p *elasticRepository) SuggestProducts(ctx context.Context, prefix string, size int) ([]Product, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Suggesting products for prefix: "+prefix)

	// 1. Build base query on name with fuzziness
	baseQuery := elastic.NewMatchQuery("name", prefix).
		Fuzziness("AUTO").
		PrefixLength(1).
		Operator("and")

	// 2. Apply filter: out_of_stock == false
	boolQuery := elastic.NewBoolQuery().
		Must(baseQuery).
		Filter(elastic.NewTermQuery("out_of_stock", false))

	// 3. Boost score based on 'sold' field
	functionScoreQuery := elastic.NewFunctionScoreQuery().
		Query(boolQuery).
		AddScoreFunc(elastic.NewFieldValueFactorFunction().
			Field("sold").
			Factor(1.5).
			Modifier("sqrt").
			Missing(0)).
		BoostMode("sum")

	// 4. Execute search
	searchResult, err := p.client.Search().
		Index("catalog").
		Query(functionScoreQuery).
		Size(size).
		Do(ctx)

	if err != nil {
		Logs.Error(ctx, "Failed to suggest products: "+err.Error())
		return nil, err
	}

	// 5. Parse results
	suggestions := []Product{}
	for _, hit := range searchResult.Hits.Hits {
		var doc productDocument
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			Logs.Error(ctx, "Unmarshal failed for suggestion: "+err.Error())
			continue
		}
		score := 0.0
		if hit.Score != nil {
			score = *hit.Score
		}

		suggestions = append(suggestions, Product{
			ID:          hit.Id,
			Name:        doc.Name,
			Description: doc.Description,
			Price:       doc.Price,
			Stock:       doc.Stock,
			Sold:        doc.Sold,
			OutOfStock:  doc.OutOfStock,
			Score:       score,
		})
	}

	Logs.Info(ctx, "Suggestions found: "+logger.IntToStr(len(suggestions)))
	return suggestions, nil
}

func (p *elasticRepository) AISuggest(ctx context.Context, query string, size int) ([]Product, error) {
	// 1. Call Python service for embedding
	embedding, err := GetEmbeddingFromPython(query, "")
	if err != nil {
		return nil, fmt.Errorf("embedding fetch failed: %w", err)
	}

	// 2. Run vector search using script_score
	searchResult, err := p.client.Search().
		Index("catalog").
		Query(elastic.NewFunctionScoreQuery().
			Query(elastic.NewMatchAllQuery()).
			AddScoreFunc(elastic.NewScriptFunction(
				elastic.NewScript("cosineSimilarity(params.query_vector, 'embedding') + 1.0").
					Param("query_vector", embedding),
			)),
		).
		Size(size).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("elasticsearch search failed: %w", err)
	}

	// 3. Parse results
	var results []Product
	for _, hit := range searchResult.Hits.Hits {
		var doc productDocument
		if err := json.Unmarshal(hit.Source, &doc); err != nil {
			continue
		}
		results = append(results, Product{
			ID:          hit.Id,
			Name:        doc.Name,
			Description: doc.Description,
			Price:       doc.Price,
			Stock:       doc.Stock,
			Sold:        doc.Sold,
			OutOfStock:  doc.OutOfStock,
		})
	}
	return results, nil
}
