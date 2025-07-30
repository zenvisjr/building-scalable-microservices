package order

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/zenvisjr/building-scalable-microservices/account"
	"github.com/zenvisjr/building-scalable-microservices/catalog"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"github.com/zenvisjr/building-scalable-microservices/mail"
	"github.com/zenvisjr/building-scalable-microservices/order/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type grpcServer struct {
	service       Service
	accountClient *account.Client
	catalogClient *catalog.Client
	mailClient    *mail.Mail
	netScan       *nats.Conn
	pb.UnimplementedOrderServiceServer
}

func ListenGRPC(s Service, accountURL, catalogURL, mailURL string, port int) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo(fmt.Sprintf("Initializing Order gRPC server on port %d", port))

	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		Logs.Error(context.Background(), "Failed to initialize Account client: "+err.Error())
		return err
	}
	Logs.LocalOnlyInfo("Connected to Account service: " + accountURL)

	catalogClient, err := catalog.NewClient(catalogURL)
	if err != nil {
		Logs.Error(context.Background(), "Failed to initialize Catalog client: "+err.Error())
		return err
	}
	Logs.LocalOnlyInfo("Connected to Catalog service: " + catalogURL)

	// mailClient, err := mail.NewMailClient(mailURL)
	// if err != nil {
	// 	Logs.Error(context.Background(), "Failed to connect to Mail gRPC: "+err.Error())
	// 	return err
	// }
	// Logs.LocalOnlyInfo("Connected to Mail service: " + mailURL)

	nc, err := nats.Connect("nats://nats:4222")
	if err != nil {
		Logs.Error(context.Background(), "Failed to connect to NATS: "+err.Error())
		return err
	}
	Logs.LocalOnlyInfo("Connected to NATS in order microservice")

	//enabling jetstream for persistent storage of order status changes for 1 day
	Logs.LocalOnlyInfo("Enabling JetStream for persistent storage of order status changes for 1 day")
	// js, err := nc.JetStream()
	// if err != nil {
	// 	log.Fatal("Failed to get JetStream context:", err)
	// }

	// _, err = js.AddStream(&nats.StreamConfig{
	// 	Name:     "ORDER_STATUS",
	// 	Subjects: []string{"order.status.changed"},
	// 	MaxAge:   24 * time.Hour, // retain messages for 1 day
	// })
	// if err != nil && !strings.Contains(err.Error(), "stream name already in use") {
	// 	log.Fatal("Failed to add stream:", err)
	// }

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		Logs.Error(context.Background(), fmt.Sprintf("Failed to bind to port %d: %v", port, err))
		accountClient.Close()
		catalogClient.Close()
		return err
	}
	Logs.LocalOnlyInfo(fmt.Sprintf("Successfully listening on port %d", port))

	server := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryLoggingInterceptor()),
	)
	pb.RegisterOrderServiceServer(server, &grpcServer{
		service:       s,
		accountClient: accountClient,
		catalogClient: catalogClient,
		// mailClient:    mailClient,
		netScan: nc,
	})

	reflection.Register(server)
	Logs.Info(context.Background(), fmt.Sprintf("Order gRPC server started on port %d", port))

	return server.Serve(conn)
}

// Take an order creation request from a client (with account ID and product list), fetch account
// & product info from other services, construct a complete order, store it, and return the full order as a response.
func (g *grpcServer) PostOrder(ctx context.Context, req *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	Logs := logger.GetGlobalLogger()

	// Logs.LocalOnlyInfo(fmt.Sprintf("PostOrder called with %d products", len(req.GetProducts())))
	// for i, p := range req.GetProducts() {
	// 	Logs.LocalOnlyInfo(fmt.Sprintf("Request Product %d: ID=%s, Quantity=%d", i, p.ProductId, p.Quantity))
	// }

	// STEP 1: Validate Account ID
	account, err := g.accountClient.GetAccount(ctx, req.GetAccountId())
	if err != nil {
		Logs.Error(ctx, "Failed to fetch account: "+err.Error())
		return nil, errors.Errorf("account not found")
	}
	Logs.LocalOnlyInfo("Account validated: " + req.GetAccountId())

	// STEP 2: Collect product IDs
	productID := []string{}
	for _, product := range req.GetProducts() {
		productID = append(productID, product.ProductId)
	}
	if len(productID) == 0 {
		Logs.Error(ctx, "Product list is empty after collecting IDs")
		return nil, errors.Errorf("Product list is empty")
	}
	Logs.LocalOnlyInfo(fmt.Sprintf("Collected %d product IDs: %v", len(productID), productID))

	// STEP 3: Fetch product details
	products, err := g.catalogClient.GetProducts(ctx, 0, 0, productID, "")
	if err != nil {
		Logs.Error(ctx, "Failed to fetch products from catalog: "+err.Error())
		return nil, errors.Errorf("Product not found")
	}
	Logs.LocalOnlyInfo(fmt.Sprintf("Catalog returned %d products", len(products)))
	for i, p := range products {
		Logs.LocalOnlyInfo(fmt.Sprintf("Catalog Product %d: ID=%s, Name=%s, Price=%f", i, p.ID, p.Name, p.Price))
	}

	// STEP 4: Merge catalog data with quantity from request
	orderedProduct := []OrderedProduct{}
	for _, p := range products {
		product := OrderedProduct{
			ProductID:   p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    0,
			Stock:       p.Stock,

		}

		for _, reqP := range req.GetProducts() {
			if p.ID == reqP.ProductId {
				//now we need to check if the quantity is available
				if reqP.Quantity > p.Stock {
					Logs.Error(ctx, "Requested quantity exceeds stock")
					continue
				}
				//now as we have stock so we will first update the db and then add it to the order
				ok, err := g.catalogClient.UpdateStockAndSold(ctx, p.ID, int(reqP.Quantity))
				if !ok || err != nil {
					Logs.Error(ctx, "Failed to update stock and sold: "+err.Error())
					return nil, err
				}
				product.Quantity = reqP.Quantity
				Logs.LocalOnlyInfo(fmt.Sprintf("Matched product %s with quantity %d", p.ID, reqP.Quantity))
				break
			}
		}

		if product.Quantity != 0 {
			orderedProduct = append(orderedProduct, product)
			Logs.LocalOnlyInfo(fmt.Sprintf("Added to final order: %s (qty: %d, price: %f)", product.Name, product.Quantity, product.Price))
		} else {
			Logs.LocalOnlyInfo(fmt.Sprintf("Skipped product %s (quantity 0)", p.ID))
		}
	}

	Logs.LocalOnlyInfo(fmt.Sprintf("Final ordered product list has %d items", len(orderedProduct)))

	// STEP 5: Create the order
	orderproto, err := g.service.PostOrder(ctx, req.GetAccountId(), orderedProduct)
	if err != nil {
		Logs.Error(ctx, "Failed to post order: "+err.Error())
		return nil, errors.Errorf("could not post order")
	}
	Logs.Info(ctx, "Order created with ID: "+orderproto.ID+" for AccountID: "+orderproto.AccountID)

	//send order confirmation email
	// go func() {
	var productLines []string
	for _, item := range orderedProduct {
		line := fmt.Sprintf("- %s x%d ($%.2f)", item.Name, item.Quantity, item.Price)
		productLines = append(productLines, line)
	}

	// err = g.mailClient.SendEmail(ctx, account.Email, "🧾 Order Confirmation", "order_confirmation", map[string]string{
	// 	"Name":  account.Name,
	// 	"Email": account.Email,
	// 	"Order": orderproto.ID,
	// 	"Total": fmt.Sprintf("$%.2f", orderproto.TotalPrice),
	// 	"Items": strings.Join(productLines, "\n"),
	// })
	// if err != nil {
	// 	Logs.Warn(ctx, "Failed to send confirmation email: "+err.Error())
	// } else {
	// 	Logs.Info(ctx, "Confirmation email sent to "+account.Email)
	// }

	emailJob := map[string]interface{}{
		"to":           account.Email,
		"subject":      "🧾 Order Confirmation",
		"templateName": "order_confirmation",
		"templateData": map[string]string{
			"Name":  account.Name,
			"Email": account.Email,
			"Order": orderproto.ID,
			"Total": fmt.Sprintf("$%.2f", orderproto.TotalPrice),
			"Items": strings.Join(productLines, "\n"),
		},
	}

	// Marshal and publish to NATS
	payload, err := json.Marshal(emailJob)
	if err != nil {
		Logs.Error(ctx, "Failed to marshal order email job: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Publishing order confirmation email to NATS")
	err = g.netScan.Publish("emails.send", payload)
	if err != nil {
		Logs.Error(ctx, "Failed to publish order email job to NATS: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Order email job published to NATS")
	// }()

	// STEP 6: Build gRPC response
	resProduct := &pb.Order{
		Id:         orderproto.ID,
		AccountId:  orderproto.AccountID,
		TotalPrice: orderproto.TotalPrice,
		CreatedAt:  timestamppb.New(orderproto.CreatedAt),
		Products:   []*pb.Order_OrderedProduct{},
	}

	// STEP 7: Simulate order status changes
	Logs.Info(ctx, "Simulating order status changes for order ID: "+orderproto.ID)
	simulateOrderStatus(ctx, orderproto.ID, g.netScan)

	for _, item := range orderproto.Products {
		resProduct.Products = append(resProduct.Products, &pb.Order_OrderedProduct{
			ProductId:   item.ProductID,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Stock:       item.Stock,

		})
	}

	return &pb.PostOrderResponse{Order: resProduct}, nil
}

func (g *grpcServer) GetOrdersForAccount(ctx context.Context, req *pb.GetOrdersForAccountRequest) (res *pb.GetOrdersForAccountResponse, err error) {
	Logs := logger.GetGlobalLogger()
	accountID := req.GetAccountId()

	Logs.Info(ctx, "Fetching orders for account ID: "+accountID)

	// Step 1: Fetch orders for the given account from the order service
	orders, err := g.service.GetOrdersByAccount(ctx, accountID)
	if err != nil {
		Logs.Error(ctx, "Failed to get orders for account ID "+accountID+": "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Fetched "+strconv.Itoa(len(orders))+" orders for account ID: "+accountID)

	// Step 2: Extract all unique product IDs from all orders to prepare for catalog lookup
	productId := map[string]bool{}
	for _, order := range orders {
		for _, product := range order.Products {
			productId[product.ProductID] = true
		}
	}
	Logs.Info(ctx, "Collected "+strconv.Itoa(len(productId))+" unique product IDs for catalog fetch")

	// Step 3: Convert the productId map to a slice for querying the catalog service
	productIds := []string{}
	for id := range productId {
		productIds = append(productIds, id)
	}

	// Step 4: Fetch full product details from the catalog service using collected IDs
	productList, err := g.catalogClient.GetProducts(ctx, 0, 0, productIds, "")
	if err != nil {
		Logs.Error(ctx, "Error getting product list from catalog service: "+err.Error())
		return nil, err
	}
	Logs.Info(ctx, "Fetched "+strconv.Itoa(len(productList))+" products from catalog service")

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
		Logs.Info(ctx, "Processing order ID: "+order.ID+" with "+strconv.Itoa(len(order.Products))+" products")

		// Step 7: For each product in this order, enrich it using catalog data
		for _, product := range order.Products {
			for _, p := range productList {
				if p.ID == product.ProductID {
					product.Name = p.Name
					product.Description = p.Description
					product.Price = p.Price
					product.Stock = p.Stock
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
				Stock:       product.Stock,
			})
		}

		// Step 9: Add the complete protobuf order to the response slice
		resProducts = append(resProducts, resProduct)
	}

	Logs.Info(ctx, "Successfully constructed response for account ID: "+accountID)
	// Step 10: Return the fully constructed response containing all enriched orders
	return &pb.GetOrdersForAccountResponse{
		Orders: resProducts,
	}, nil
}

func simulateOrderStatus(ctx context.Context, orderID string, nc *nats.Conn) {
	statuses := []string{"✅ Confirmed", "📦 Packed", "🚚 Shipped", "📦 Delivered"}
	Logs := logger.GetGlobalLogger()

	go func() {
		Logs.Info(ctx, "Simulating order status changes for order ID: "+orderID)
		for _, status := range statuses {
			localCtx := context.Background()
			Logs := logger.GetGlobalLogger()
			Logs.Info(localCtx, "Simulating status: "+status+" for "+orderID)
			time.Sleep(15 * time.Second)
			payload := OrderStatusUpdate{
				OrderID:   orderID,
				Status:    status,
				UpdatedAt: time.Now().Format(time.RFC3339),
			}
			data, err := json.Marshal(payload)
			if err != nil {
				Logs.Error(localCtx, "Failed to marshal order status update: "+err.Error())
				return
			}
			Logs.Info(localCtx, "Publishing order status update to NATS")
			// nc.Publish("order.status.changed", data)
			// js, _ := nc.JetStream() // reuse connection

			nc.Publish("order.status.changed", data)
		}
	}()
}

// its auto generated in graphql but we needed here so just copy pasting RELAX BITCH
type OrderStatusUpdate struct {
	OrderID   string `json:"orderId"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updatedAt"`
}
