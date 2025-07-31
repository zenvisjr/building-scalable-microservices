package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zenvisjr/building-scalable-microservices/logger"
)

type embeddingRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type embeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func GetEmbeddingFromPython(name, description string) ([]float64, error) {
	payload := embeddingRequest{
		Name:        name,
		Description: description,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	Logs := logger.GetGlobalLogger()
	Logs.Info(context.Background(), "Inside GetEmbeddingFromPython: "+name)
	resp, err := http.Post("http://embed_service:5005/embed", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	Logs.Info(context.Background(), "got response from embed_service: "+name)

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding service error: %s", msg)
	}
	Logs.Info(context.Background(), "got response from embed_service: "+name)
	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	Logs.Info(context.Background(), " response sent from embed_service: "+name)

	return result.Embedding, nil
}

func (p *elasticRepository) AddEmbeddingToProduct(product Product) {
	Logs := logger.GetGlobalLogger()
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	Logs.Info(ctx, "Started embedding for product: " + product.ID)

	embedding, err := GetEmbeddingFromPython(product.Name, product.Description)
	if err != nil {
		Logs.Error(ctx, "Embedding failed for "+product.ID+": "+err.Error())
		return
	}

	Logs.Info(ctx, "Going to update the product with embedding: "+product.ID)

	_, err = p.client.Update().
		Index("catalog").
		Id(product.ID).
		Doc(map[string]interface{}{"embedding": embedding}).
		Do(ctx)

	if err != nil {
		Logs.Error(ctx, "Failed to update product with embedding: "+err.Error())
	} else {
		Logs.Info(ctx, "Embedding added to product: " + product.ID)
	}
}
