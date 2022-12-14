package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	pb "github.com/spliffone/grpc-playground/basics/go/proto"

	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func initGRPCServerHTTP2() int {
	lis, err := net.Listen("tcp", ":0")

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterProductInfoServer(s, &productService{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return lis.Addr().(*net.TCPAddr).Port
}

func TestServer_AddProduct(t *testing.T) {
	port := initGRPCServerHTTP2()

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewProductInfoClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.AddProduct(ctx, &pb.Product{Name: "Product",
		Description: "ProductDescription", Price: float32(700.0)})
	if err != nil {
		t.Fatalf("Could not add product: %v", err)
	}
	if r.Value == "" {
		t.Errorf("Invalid Product ID %s", r.Value)
	}

	log.Printf("Res %s", r.Value)
}
