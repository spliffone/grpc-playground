package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"

	pb "github.com/spliffone/grpc-playground/observability/go/proto"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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
	host = flag.String("host", LookupEnvOrString("PRODUCT_INFO_SERVER", "localhost"), "The product info server host")
	port = flag.Int("port", LookupEnvOrInt("PRODUCT_INFO_SERVER_PORT", 50051), "The server port")
)

func main() {
	flag.Parse()

	serverAddr := net.JoinHostPort(*host, strconv.Itoa(*port))

	// Call the initTracing function and initialize the Jaeger exporter
	// instance and register with trace.
	tp, err := InitTraceProvider("productinfo.client", "test", "http://localhost:14268/api/traces")
	if err != nil {
		log.Fatalln("trace provider init failed", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	// Register stats and trace exporters to export the collected data.
	// Here we will add PrintExporter, which logs exported data to the console.
	// This is only for demonstration purposes. Normally it is not recommended to
	// log all production loads.
	view.RegisterExporter(&exporter.PrintExporter{})
	if err := view.Register(ocgrpc.DefaultClientViews...); err != nil {
		log.Fatal(err)
	}

	opts := []grpc.DialOption{
		// Register the views to collect server request count. These are the
		// predefined default service views that collect received bytes per RPC,
		// sent bytes per RPC, latency per RPC, and completed RPC. We can write
		// our own views to collect data.
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewProductInfoClient(conn)

	callServer(c, tp.Tracer("component-main"))
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func LookupEnvOrInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

// prepareContext with time and timeout
func prepareContext(tr trace.Tracer) (context.Context, context.CancelFunc, trace.Span) {
	// Setup deadline
	clientDeadline := time.Now().Add(time.Duration(timeout * time.Second))
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	// Start new span with span name and context.
	ctx, span := tr.Start(ctx, "productinfo.ProductInfoClient")

	// Create Client Metadata which is send to the server
	return metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
	)), cancel, span
}

func callServer(c pb.ProductInfoClient, tr trace.Tracer) {

	// create a new product
	r := addProduct(c, &pb.Product{
		Name:        "Apple iPhone 11",
		Description: "Meet Apple iPhone 11.",
		Price:       float32(1000.0)},
		tr)
	// get the new product
	getProduct(c, r.Value, tr)
	// search for a product
	searchProducts(c, tr)
	// get an extended error since the product ID is invalid
	getProduct(c, "-1", tr)
}

func addProduct(c pb.ProductInfoClient, newProduct *pb.Product, tr trace.Tracer) *pb.ProductID {
	ctx, cancel, span := prepareContext(tr)
	defer cancel()
	// Stop the span when everything is done.
	defer span.End()

	var responseHeader, trailer metadata.MD
	r, err := c.AddProduct(ctx,
		newProduct,
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

func getProduct(c pb.ProductInfoClient, v string, tr trace.Tracer) {
	ctx, cancel, span := prepareContext(tr)
	defer cancel()
	// Stop the span when everything is done.
	defer span.End()

    span.SetAttributes(attribute.Key("ProductID").String(v))
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

func searchProducts(c pb.ProductInfoClient, tr trace.Tracer) {
	ctx, cancel, span := prepareContext(tr)
	defer cancel()
	// Stop the span when everything is done.
	defer span.End()

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
