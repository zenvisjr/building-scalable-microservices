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

	// STEP 1: Validate Account ID by calling Account microservice
	_, err := g.accountClient.GetAccount(ctx, req.GetAccountId())
	if err != nil {
		log.Println("Error getting account", err)
		return nil, errors.Errorf("account not found")
	}

	// STEP 2: Collect all product IDs from request
	productID := []string{}
	for _, product := range req.GetProducts() {
		productID = append(productID, product.ProductId)
	}

	// STEP 3: Fetch full product details from Catalog microservice
	products, err := g.catalogClient.GetProducts(ctx, 0, 0, productID, "")
	if err != nil {
		log.Println("Error getting product", err)
		return nil, errors.Errorf("Product not found")
	}

	// STEP 4: Merge the product details with the quantity info from request
	// and create a list of product that we want to order
	orderedProduct := []OrderedProduct{}
	for _, p := range products {
		product := OrderedProduct{
			ProductID:   p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    0,
		}

		for _, resp := range req.GetProducts() {
			if product.ProductID == resp.ProductId {
				product.Quantity = resp.Quantity
				break
			}
		}

		if product.Quantity != 0 {
			orderedProduct = append(orderedProduct, product)
		}
	}

	// STEP 5: Create the order in the Order Service
	orderproto, err := g.service.PostOrder(ctx, req.GetAccountId(), orderedProduct)
	if err != nil {
		log.Println("Error posting order: ", err)
		return nil, errors.Errorf("could not post order:", err)
	}

	//now we need to create Order that have OrderedProducts and send back as response

	// STEP 6: Map the created order to protobuf format
	resProduct := &pb.Order{
		Id:         orderproto.ID,
		AccountId:  orderproto.AccountID,
		TotalPrice: orderproto.TotalPrice,
		CreatedAt:  timestamppb.New(orderproto.CreatedAt),
		Products:   []*pb.Order_OrderedProduct{},
	}
	// STEP 7: Add all OrderedProducts to response
	for _, item := range orderproto.Products {
		resProduct.Products = append(resProduct.Products, &pb.Order_OrderedProduct{
			ProductId:   item.ProductID,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})

	}

	// STEP 8: Return response with created order
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
	for id, _ := range productId {
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
