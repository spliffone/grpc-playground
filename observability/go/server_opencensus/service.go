package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/gogo/googleapis/google/rpc"
	pb "github.com/spliffone/grpc-playground/observability/go/proto"
	"go.opencensus.io/trace"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// service is used to implement ProductInfo.
type productService struct {
	pb.UnimplementedProductInfoServer
	mu         sync.Mutex
	productMap map[string]*pb.Product
}

// AddProduct implements ProductInfo.AddProduct
func (s *productService) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	// Start new span with span name and context
	ctx, span := trace.StartSpan(ctx, "productinfo.AddProduct")
	// Stop the span when everything is done.
	defer span.End()
	log.Printf("AddProduct: %v", in)

	out, err := uuid.NewV4()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error while generating Product ID: %s", err)
	}

	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}

	// Prevent concurrent write map access
	s.mu.Lock()
	defer s.mu.Unlock()
	s.productMap[in.Id] = in

	return &pb.ProductID{Value: in.Id}, status.New(codes.OK, "").Err()
}

// GetProduct implements ProductInfo.GetProduct
func (s *productService) GetProduct(ctx context.Context, in *pb.ProductID) (*pb.Product, error) {
	// Start new span with span name and context
	ctx, span := trace.StartSpan(ctx, "productinfo.GetProduct")
	// Stop the span when everything is done.
	defer span.End()
	log.Printf("GetProduct: %v", in)

	if in.Value == "-1" {
		// see https://jbrandhorst.com/post/grpc-errors/
		return nil, invalidID(in.Value, "ProductID.Value")
	} else {
		value, exists := s.productMap[in.Value]
		if exists {
			return value, status.New(codes.OK, "").Err()
		}
		return nil, status.Error(codes.NotFound, "Product does not exist.")
	}
}

// SearchOrders implements ProductInfo.SearchProducts
func (s *productService) SearchOrders(searchQuery *pb.SearchQuery, stream pb.ProductInfo_SearchProductsServer) error {
	for key, product := range s.productMap {
		log.Print(key, product)

		if strings.Contains(product.Description, searchQuery.Value) {
			// Send the matching orders in a stream
			err := stream.Send(product)
			if err != nil {
				return fmt.Errorf("error sending message to stream: %v", err)
			}
			log.Print("Matching Product Found : " + key)
			break
		}

	}
	return nil
}

// invalidID build invalid product ID error
func invalidID(id string, field string) error {
	log.Printf("Product ID is invalid! -> Received Product ID %s", id)

	errorStatus := status.New(codes.InvalidArgument, "Invalid information received")
	ds, err := errorStatus.WithDetails(&rpc.BadRequest{
		FieldViolations: []*rpc.BadRequest_FieldViolation{{
			Field:       field,
			Description: fmt.Sprintf("Product ID received is not valid %s", id),
		}},
	})
	if err != nil {
		return errorStatus.Err()
	}
	return ds.Err()
}
