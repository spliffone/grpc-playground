package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	mock_productinfo "github.com/spliffone/grpc-playground/basics/go/mocks"
	pb "github.com/spliffone/grpc-playground/basics/go/proto"
	"google.golang.org/protobuf/proto"
)

const (
	name        = "Product Name"
	description = "Product Description"
	price       = float32(99.99)
)

// rpcMsg implements the gomock.Matcher interface
type rpcMsg struct {
	msg proto.Message
}

func (r *rpcMsg) Matches(msg interface{}) bool {
	m, ok := msg.(proto.Message)
	if !ok {
		return false
	}
	return proto.Equal(m, r.msg)
}

func (r *rpcMsg) String() string {
	return fmt.Sprintf("is %s", r.msg)
}

func TestAddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock client
	mockClient := mock_productinfo.NewMockProductInfoClient(ctrl)

	req := &pb.Product{Name: name, Description: description, Price: price}

	mockClient.
		EXPECT().AddProduct(gomock.Any(), &rpcMsg{msg: req}, gomock.Any()).
		Return(&pb.ProductID{Value: "ABC123" + name}, nil)

	testAddProduct(t, mockClient, req)
}

func testAddProduct(t *testing.T, mockClient pb.ProductInfoClient, req *pb.Product) {
	r := addProduct(mockClient, req)
	// test and verify response.
	if !strings.HasSuffix(r.Value, name) {
		t.Errorf("The AddProduct response ProductID.Value shall end with '%s' but the value is %s", name, r.Value)
	}
}
