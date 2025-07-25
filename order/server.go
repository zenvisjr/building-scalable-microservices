package order

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/pkg/errors"
	"github.com/zenvisjr/building-scalable-microservices/account"
	"github.com/zenvisjr/building-scalable-microservices/catalog"
	"github.com/zenvisjr/building-scalable-microservices/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type grpcServer struct {
	service       Service
	accountClient *account.Client
	catalogClient *catalog.Client
	pb.UnimplementedOrderServiceServer
}

func ListenGRPC(s Service, accountURL, catalogURL string, port int) error {
	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		return err
	}

	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		return err
	}

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		accountClient.Close()
		catalogClient.Close()
		return err
	}
	server := grpc.NewServer()
	pb.RegisterOrderServiceServer(server, &grpcServer{
		service:         s,
		accountClient:   accountClient,
		catalogClient:   catalogClient,
	})

	reflection.Register(server)

	return server.Serve(conn)
}

// Take an order creation request from a client (with account ID and product list), fetch account
// & product info from other services, construct a complete order, store it, and return the full order as a response.
func (g *grpcServer) PostOrder(ctx context.Context, req *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	log.Printf("🔥 gRPC PostOrder called with %d products", len(req.GetProducts()))
	
	// Log each product in the request
	for i, p := range req.GetProducts() {
		log.Printf("🔥 Request Product %d: ID=%s, Quantity=%d", i, p.ProductId, p.Quantity)
	}

	// STEP 1: Validate Account ID by calling Account microservice
	_, err := g.accountClient.GetAccount(ctx, req.GetAccountId())
	if err != nil {
		log.Println("❌ Error getting account", err)
		return nil, errors.Errorf("account not found")
	}
	log.Printf("✅ Account validated: %s", req.GetAccountId())

	// STEP 2: Collect all product IDs from request
	productID := []string{}
	for _, product := range req.GetProducts() {
		productID = append(productID, product.ProductId)
	}
	if len(productID) == 0 {
		log.Printf("❌ Product list is empty after collection!")
		return nil, errors.Errorf("Product list is empty")
	}
	log.Printf("🔥 Collected %d product IDs: %v", len(productID), productID)

	// STEP 3: Fetch full product details from Catalog microservice
	products, err := g.catalogClient.GetProducts(ctx, 0, 0, productID, "")
	if err != nil {
		log.Printf("❌ Error getting products from catalog: %v", err)
		return nil, errors.Errorf("Product not found")
	}
	log.Printf("🔥 Catalog returned %d products", len(products))
	
	// Log each product from catalog
	for i, p := range products {
		log.Printf("🔥 Catalog Product %d: ID=%s, Name=%s, Price=%f", i, p.ID, p.Name, p.Price)
	}

	// STEP 4: Merge the product details with the quantity info from request
	orderedProduct := []OrderedProduct{}
	for _, p := range products {
		product := OrderedProduct{
			ProductID:   p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    0,
		}

		// Find matching quantity from request
		for _, resp := range req.GetProducts() {
			if p.ID == resp.ProductId {
				product.Quantity = resp.Quantity
				log.Printf("🔥 Matched product %s with quantity %d", p.ID, resp.Quantity)
				break
			}
		}

		if product.Quantity != 0 {
			orderedProduct = append(orderedProduct, product)
			log.Printf("✅ Added to final order: %s (qty: %d, price: %f)", product.Name, product.Quantity, product.Price)
		} else {
			log.Printf("❌ Product %s skipped - quantity is 0", p.ID)
		}
	}

	log.Printf("🔥 Final orderedProduct list has %d items", len(orderedProduct))

	// STEP 5: Create the order in the Order Service
	orderproto, err := g.service.PostOrder(ctx, req.GetAccountId(), orderedProduct)
	if err != nil {
		log.Println("❌ Error posting order: ", err)
		return nil, errors.Errorf("could not post order:", err)
	}

	// Rest of your existing code...
	resProduct := &pb.Order{
		Id:         orderproto.ID,
		AccountId:  orderproto.AccountID,
		TotalPrice: orderproto.TotalPrice,
		CreatedAt:  timestamppb.New(orderproto.CreatedAt),
		Products:   []*pb.Order_OrderedProduct{},
	}
	
	for _, item := range orderproto.Products {
		resProduct.Products = append(resProduct.Products, &pb.Order_OrderedProduct{
			ProductId:   item.ProductID,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	return &pb.PostOrderResponse{
		Order: resProduct,
	}, nil
}

func (g *grpcServer) GetOrdersForAccount(ctx context.Context, req *pb.GetOrdersForAccountRequest) (res *pb.GetOrdersForAccountResponse, err error) {
	// Step 1: Fetch orders for the given account from the order service
	orders, err := g.service.GetOrdersByAccount(ctx, req.GetAccountId())
	if err != nil {
		return nil, err
	}

	// Step 2: Extract all unique product IDs from all orders to prepare for catalog lookup
	productId := map[string]bool{}
	for _, order := range orders {
		for _, product := range order.Products {
			productId[product.ProductID] = true
		}
	}

	// Step 3: Convert the productId map to a slice for querying the catalog service
	productIds := []string{}
	for id := range productId {
		productIds = append(productIds, id)
	}
	// Step 4: Fetch full product details from the catalog service using collected IDs
	productList, err := g.catalogClient.GetProducts(ctx, 0, 0, productIds, "")
	if err != nil {
		log.Println("Error getting product list:", err)
		return nil, err
	}

	// Step 5: Initialize the final response list of protobuf orders
	resProducts := []*pb.Order{}

	// Step 6: Loop through all domain orders and map them to protobuf orders
	for _, order := range orders {
		resProduct := &pb.Order{
			Id:         order.ID,
			AccountId:  order.AccountID,
			TotalPrice: order.TotalPrice,
			CreatedAt:  timestamppb.New(order.CreatedAt),
			Products:   []*pb.Order_OrderedProduct{},
		}
		// Step 7: For each product in this order, enrich it using catalog data
		for _, product := range order.Products {
			for _, p := range productList {
				if p.ID == product.ProductID {
					// Fill in missing fields like name, description, price from catalog
					product.Name = p.Name
					product.Description = p.Description
					product.Price = p.Price
					break

				}
			}
			// Step 8: Add the enriched product to protobuf order's product list
			resProduct.Products = append(resProduct.Products, &pb.Order_OrderedProduct{
				ProductId:   product.ProductID,
				Name:        product.Name,
				Description: product.Description,
				Price:       product.Price,
				Quantity:    product.Quantity,
			})
		}
		// Step 9: Add the complete protobuf order to the response slice
		resProducts = append(resProducts, resProduct)
	}

	// Step 10: Return the fully constructed response containing all enriched orders
	return &pb.GetOrdersForAccountResponse{
		Orders: resProducts,
	}, nil
}
