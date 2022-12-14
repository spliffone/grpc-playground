package main

import (
	"context"
	"fmt"

	pb "github.com/spliffone/grpc-playground/resolver/go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"log"
)

func main() {
	addr := fmt.Sprintf("%s:///%s", exampleScheme, exampleServiceName)

	conn, err := grpc.Dial(addr,
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`, "round_robin")), // This sets the initial balancing policy.
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewProductInfoClient(conn)

	r, err := c.AddProduct(context.Background(), &pb.Product{})
	if err != nil {
		log.Fatalf("did not create product %v", err)
	}
	log.Printf("Product ID: %s added successfully", r.Value)
}
