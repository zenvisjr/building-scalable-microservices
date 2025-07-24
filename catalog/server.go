package catalog

import (
	"context"
	"fmt"
	"net"

	"github.com/zenvisjr/building-scalable-microservices/catalog/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	service Service
	pb.UnimplementedCatalogServiceServer
}

func ListenGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	//register service

	pb.RegisterCatalogServiceServer(server, &grpcServer{service: s})

	//register reflection
	reflection.Register(server)

	return server.Serve(lis)

}

func (g *grpcServer) PostProduct(ctx context.Context, req *pb.PostProductRequest) (*pb.PostProductResponse, error) {
	product, err := g.service.PostProduct(ctx, req.GetName(), req.GetDescription(), req.GetPrice())
	if err != nil {
		return nil, err
	}
	response := &pb.PostProductResponse{
		Product: &pb.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		},
	}
	return response, nil
}

func (g *grpcServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, err := g.service.GetProduct(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	response := &pb.GetProductResponse{
		Product: &pb.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		},
	}
	return response, nil
}

func (g *grpcServer) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	var (
		err      error
		resp     []Product
	)

	if len(req.GetQuery()) > 0 {
		resp, err = g.service.SearchProducts(ctx, req.GetQuery(), req.GetSkip(), req.GetTake())
	} else if len(req.GetIds()) != 0 {
		resp, err = g.service.GetProductsByIDs(ctx, req.GetIds())
	} else {
		resp, err = g.service.GetProducts(ctx, req.GetSkip(), req.GetTake())
	}

	if err != nil {
		return nil, err
	}

	products := make([]*pb.Product, len(resp))
	for i, p := range resp {
		products[i] = &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		}
	}
	response := &pb.GetProductsResponse{
		Products: products,
	}
	return response, nil
}
