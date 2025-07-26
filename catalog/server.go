package catalog

import (
	"context"
	"fmt"
	"net"

	"github.com/zenvisjr/building-scalable-microservices/catalog/pb"
	"github.com/zenvisjr/building-scalable-microservices/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	service Service
	pb.UnimplementedCatalogServiceServer
}

func ListenGRPC(s Service, port int) error {
	Logs := logger.GetGlobalLogger()
	Logs.LocalOnlyInfo("Initializing gRPC listener for catalog microservice on " + fmt.Sprintf(":%d", port))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		Logs.Error(context.Background(), "Failed to bind to port: "+err.Error())
		return err
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryLoggingInterceptor()),
	)
	pb.RegisterCatalogServiceServer(server, &grpcServer{service: s})
	reflection.Register(server)

	Logs.LocalOnlyInfo("Catalog gRPC server started on port " + fmt.Sprintf("%d", port))
	return server.Serve(lis)
}

func (g *grpcServer) PostProduct(ctx context.Context, req *pb.PostProductRequest) (*pb.PostProductResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Received PostProduct request")

	product, err := g.service.PostProduct(ctx, req.GetName(), req.GetDescription(), req.GetPrice())
	if err != nil {
		Logs.Error(ctx, "PostProduct failed: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Product created successfully: "+product.ID)

	return &pb.PostProductResponse{
		Product: &pb.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		},
	}, nil
}

func (g *grpcServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	Logs := logger.GetGlobalLogger()
	Logs.Info(ctx, "Received GetProduct request for ID: "+req.GetId())

	product, err := g.service.GetProduct(ctx, req.GetId())
	if err != nil {
		Logs.Error(ctx, "GetProduct failed: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, "Product fetched: "+product.ID)

	return &pb.GetProductResponse{
		Product: &pb.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		},
	}, nil
}

func (g *grpcServer) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	Logs := logger.GetGlobalLogger()

	Logs.Info(ctx, "Received GetProducts request")

	var (
		err  error
		resp []Product
	)

	switch {
	case len(req.GetQuery()) > 0:
		Logs.Info(ctx, "Searching products with query: "+req.GetQuery())
		resp, err = g.service.SearchProducts(ctx, req.GetQuery(), req.GetSkip(), req.GetTake())

	case len(req.GetIds()) > 0:
		Logs.Info(ctx, "Fetching products by IDs")
		resp, err = g.service.GetProductsByIDs(ctx, req.GetIds())

	default:
		Logs.Info(ctx, "Fetching all products with pagination")
		resp, err = g.service.GetProducts(ctx, req.GetSkip(), req.GetTake())
	}

	if err != nil {
		Logs.Error(ctx, "GetProducts failed: "+err.Error())
		return nil, err
	}

	Logs.Info(ctx, fmt.Sprintf("Fetched %d products", len(resp)))

	products := make([]*pb.Product, len(resp))
	for i, p := range resp {
		products[i] = &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}
	}

	return &pb.GetProductsResponse{
		Products: products,
	}, nil
}
