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

	pb "github.com/spliffone/grpc-playground/observability/go/proto"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/zpages"
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

	// Starts a z-Pages server. An HTTP endpoint starts with the context of
	// /debug in port 8081 for metrics visualization.
	go func() {
		mux := http.NewServeMux()
		zpages.Handle(mux, "/debug")
		log.Println("Starting HTTP zpages listener on port 8081")
		log.Fatal(http.ListenAndServe("127.0.0.1:8081", mux))
	}()
	// Register stat exporters to export the collected data.
	// Here we add PrintExporter and it logs exported data to the console.
	// This is only for demonstration purposes; normally it’s not recommended
	// that you log all production loads.
	view.RegisterExporter(&exporter.PrintExporter{})
	// Register the views to collect the server request count.
	// These are the predefined default service views that collect received bytes per RPC,
	// sent bytes per RPC, latency per RPC, and completed RPC.
	// We can write our own views to collect data.
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Set up a new server with the OpenCensus stats handler to enable stats and tracing.
	s := grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	service := &productService{}
	pb.RegisterProductInfoServer(s, service)

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
