package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"grpc-demo/api"
	"grpc-demo/internal/service"
	"log"
	"net"
	"net/http"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	lis, err := net.Listen("tcp", ":22223")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	// 注册服务
	api.RegisterGreeterHandlerServer(ctx, mux, service.NewHelloService())

	srv := &http.Server{
		Addr:    "127.0.0.1:22223",
		Handler: mux,
	}

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}