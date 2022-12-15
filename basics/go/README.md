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

## Getting Started

Run gRPC Server:

```bash
go run $(ls -1 server/*.go | grep -v _test.go)
```

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

## Load testing

The following code show how to run load tests with [ghz - gRPC benchmarking and load testing tool](https://ghz.sh/).

```bash
curl -v -L --fail https://github.com/bojand/ghz/releases/download/v0.111.0/ghz-windows-x86_64.zip -o ghz-windows-x86_64.zip
unzip ghz-windows-x86_64.zip

echo "Run load test"
./ghz --insecure\
 --async \
 --proto=product.proto \
 --call=product.ProductInfo/addProduct \
 -d '{"name":"Product {{.RequestNumber}}", "description":"Test Description {{.RequestNumber}}", "price": 17.13}' \
 localhost:50051
```

## Build and Run Docker Container

Build the basic server:

```bash
docker build -f ./server/Dockerfile -t grpc-productinfo-basic-server:latest .
```

Run the basic server (Windows git bash):

```bash
winpty docker run -it grpc-productinfo-basic-server:latest
```

Run client and server:

```bash
docker build -f ./server/Dockerfile -t grpc-productinfo-basic-server:latest .
docker build -f ./client/Dockerfile -t grpc-productinfo-basic-client:latest .

docker network create product-net
winpty docker run -it --network=product-net --hostname=productinfo -p 50051:50051 grpc-productinfo-basic-server:latest
winpty docker run -it --network=product-net --hostname=client --env PRODUCT_INFO_SERVER=productinfo grpc-productinfo-basic-client:latest
```

