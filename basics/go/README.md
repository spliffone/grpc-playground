# Basics

Basic client server example.

Server Features:

- Simple logging interceptor for stream and unary requests
- Enrich response interceptor

Client Features:

- Request timeout
- Pass additional headers to each RPC request
- Enhanced RPC error status parsing

Test:

- Test client using [GoMock](https://github.com/golang/mock)
- Test service starting a local gRPC server

## Re-Generate Code

- proto_path - is there to specify a proto root source directory
- go_out - is the directory where you want the compiler to write to protocol buffer output
- go_opt=paths=source_relative - the output file is placed in the same relative directory as the input file.

```bash
echo Generating client and server code
protoc --proto_path=$(pwd)/../.. \
  --go_out=proto\
  --go_opt=paths=source_relative\
  --go-grpc_out=proto\
  --go-grpc_opt=paths=source_relative\
   $(pwd)/../../product.proto
```

## Re-Generate [GoMocks](https://github.com/golang/mock)

```bash
echo Install GoMock 
go install github.com/golang/mock/mockgen@v1.6.0
mkdir mocks
mockgen -source=proto/product_grpc.pb.go ProductInfoClient > mocks/productinfo_mock.go
```
