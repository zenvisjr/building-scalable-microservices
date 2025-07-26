package logger

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

type ctxKey string

const methodCtxKey ctxKey = "grpc_method_name"

// Inject method name into context for downstream logging
func UnaryLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Logs := GetGlobalLogger()
		log.Println("Intercepting gRPC call: "+info.FullMethod)
		ctx = context.WithValue(ctx, methodCtxKey, info.FullMethod)
		return handler(ctx, req)
	}
}

// Retrieve method name from context
func GetMethodFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(methodCtxKey).(string); ok {
		return val
	}
	return ""
}