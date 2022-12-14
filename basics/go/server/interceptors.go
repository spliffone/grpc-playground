package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
)

// Server - Unary Interceptor
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Preprocessing logic

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}

	// Gets info about the current RPC call by examining the args passed in
	log.Printf("[Server Interceptor] %s Meta: %v", info.FullMethod, md)

	// Invoking the handler to complete the normal execution of a unary RPC.
	m, err := handler(ctx, req)

	// Post processing logic
	log.Printf(" Post Proc Message : %s", m)

	return m, err
}

// Server - Unary Interceptor
func enrichResponseInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Invoking the handler to complete the normal execution of a unary RPC.
	m, err := handler(ctx, req)

	// Create and send headers
	grpc.SendHeader(ctx, metadata.Pairs("header-key", "header-val"))
	grpc.SetTrailer(ctx, metadata.Pairs("trailer-key", "trailer-val"))

	return m, err
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("[Server Stream Interceptor Wrapper] Receive a message (Type: %T) at %s",
		m, time.Now().Format(time.RFC3339))
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("[Server Stream Interceptor Wrapper] Send a message (Type: %T) at %v",
		m, time.Now().Format(time.RFC3339))

	// Create and send headers
	w.ServerStream.SendHeader(metadata.Pairs("header-key", "header-val"))
	w.ServerStream.SetTrailer(metadata.Pairs("trailer-key", "trailer-val"))
	return w.ServerStream.SendMsg(m)
}

func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{s}
}

func orderServerStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println("[Server Stream Interceptor] ", info.FullMethod)
	err := handler(srv, newWrappedStream(ss))
	if err != nil {
		log.Printf("RPC failed with error %v", err)
	}

	return err
}
