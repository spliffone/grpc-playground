package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	pb "github.com/spliffone/grpc-playground/observability/go/proto"
	"google.golang.org/grpc"
)

var (
	port                = flag.Int("port", 50051, "The server port")
	reg                 = prometheus.NewRegistry()
	grpcMetrics         = grpc_prometheus.NewServerMetrics()
	customMetricCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "product_mgt_server_handle_count",
		Help: "Total number of RPCs handled on the server.",
	}, []string{"name"})
)

func init() {
	// Registers standard server metrics and custom metrics collector to the
	// registry created in step 2.
	reg.MustRegister(grpcMetrics, customMetricCounter)
}

func main() {
	flag.Parse()

	// Start your http server for prometheus.
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":8082", nil)
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Set up a new server with the Prometheus stats handler.
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpcMetrics.UnaryServerInterceptor()),
	)
	service := &productService{}
	pb.RegisterProductInfoServer(s, service)
	// Initializes all standard metrics.
	grpcMetrics.InitializeMetrics(s)

	// Run gRPC server
	go func() {
		log.Printf("Starting gRPC listener on port %d", *port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Catch CTRL+C signals and gracefully shutting down
	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+C
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	<-signalChan
	log.Println("os.Interrupt - shutting down...")
	go func() {
		<-signalChan
		log.Fatalln("os.Kill - terminating...")
	}()

	s.GracefulStop()
	log.Println("gracefully stopped")

	defer os.Exit(0)
}
