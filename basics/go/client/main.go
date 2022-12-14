package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strconv"

	pb "github.com/spliffone/grpc-playground/basics/go/proto"
	"google.golang.org/genproto/googleapis/rpc/errdetails"

	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"time"

	"google.golang.org/grpc"
)

const (
	timeout = 2
)

var (
	port = flag.Int("port", 50051, "The server port (will be ignored in case we use the custom resolver)")
)

func main() {
	flag.Parse()

	serverAddr := net.JoinHostPort("localhost", strconv.Itoa(*port))

	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(loggingInterceptor),
		grpc.WithStreamInterceptor(clientStreamInterceptor),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewProductInfoClient(conn)

	callServer(c)
}

// prepareContext with time and timeout
func prepareContext() (context.Context, context.CancelFunc) {
	// Setup deadline
	clientDeadline := time.Now().Add(time.Duration(timeout * time.Second))
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)

	// Create Client Metadata which is send to the server
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
	)), cancel
}

func callServer(c pb.ProductInfoClient) {

	// create a new product
	r := addProduct(c)
	// get the new product
	getProduct(c, r.Value)
	// search for a product
	searchProducts(c)
	// get an extended error since the product ID is invalid
	getProduct(c, "-1")
}

func addProduct(c pb.ProductInfoClient) *pb.ProductID {
	ctx, cancel := prepareContext()
	defer cancel()

	var responseHeader, trailer metadata.MD
	r, err := c.AddProduct(ctx, &pb.Product{
		Name:        "Apple iPhone 11",
		Description: "Meet Apple iPhone 11.",
		Price:       float32(1000.0)},
		grpc.Header(&responseHeader),
		grpc.Trailer(&trailer))
	if err != nil {
		got := status.Code(err)
		log.Printf("Error Occurred -> addOrder : , %v:", got)
		log.Fatalf("Could not add product: %v", err)
	}

	log.Printf("Product ID: %s added successfully headers: %v, trailers: %v", r.Value, responseHeader, trailer)
	return r
}

func getProduct(c pb.ProductInfoClient, v string) {
	ctx, cancel := prepareContext()
	defer cancel()

	product, err := c.GetProduct(ctx, &pb.ProductID{Value: v})
	if err != nil {
		errorCode := status.Code(err)
		errorStatus := status.Convert(err)
		log.Printf("Get Product Error : %s", errorCode)
		for _, detail := range errorStatus.Details() {
			log.Printf("%v", reflect.TypeOf(detail).String())

			switch t := detail.(type) {
			case *errdetails.BadRequest:
				fmt.Println("Your request was rejected by the server.")
				for _, violation := range t.GetFieldViolations() {
					fmt.Printf("The %q field was wrong:\n", violation.GetField())
					fmt.Printf("\t%s\n", violation.GetDescription())
				}
			default:
				log.Printf("Unexpected error type: %s", t)
			}
		}

	}
	log.Printf("Product: %s", product.String())
}

func searchProducts(c pb.ProductInfoClient) {
	ctx, cancel := prepareContext()
	defer cancel()

	searchStream, err := c.SearchProducts(ctx, &pb.SearchQuery{Value: "Apple"})
	if err != nil {
		log.Fatalf("Could not get products: %v", err)
	}
StreamLoop:
	for {
		searchProduct, err := searchStream.Recv()
		if err != nil {
			switch err {
			case io.EOF:
				break StreamLoop
			default:
				log.Printf("Unhandled error %s", err)
				break StreamLoop
			}
		}

		// handle other possible errors
		log.Print("Search Result : ", searchProduct)
	}
}
