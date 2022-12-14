package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	pb "github.com/spliffone/grpc-playground/basics/go/proto"
	"google.golang.org/grpc"
)

type wrappedStream struct {
	grpc.ServerStream
}

var (
	port = flag.Int("port", 50051, "The server port")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	interceptors := []grpc.UnaryServerInterceptor{
		loggingInterceptor,
		enrichResponseInterceptor,
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)),
		grpc.StreamInterceptor(orderServerStreamInterceptor))

	service := &productService{}
	pb.RegisterProductInfoServer(s, service)
	log.Printf("Starting gRPC listener on port %d", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
