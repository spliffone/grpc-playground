package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

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
