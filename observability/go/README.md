# Observability

Observability client server example (see [opencensus - Installin grpc for go](https://opencensus.io/guides/grpc/go/)#installing-grpc-for-go).

Server Features (see server_opencensus/opentelemetry):

- Host GRPC metrics `http://localhost:8081/debug/rpcz`
- Host traces `http://localhost:8081/debug/tracez`

Server Features (see server_prometheus):

- Host prometheus metrics `http://localhost:8082/metrics`

More interesting tools and libs are listed under [awesome-grpc](https://github.com/grpc-ecosystem/awesome-grpc).

## Getting Started

Run gRPC Server:

```bash
echo Verbosity means how many times any single info message should print every five minutes. The verbosity is set to 0 by default.
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
echo Sets log severity level to info. All the informational messages will be printed.
export GRPC_GO_LOG_SEVERITY_LEVEL=info
go run $(ls -1 server/*.go | grep -v _test.go)
```

Run Jaeger and check the traces under `http://localhost:16686/search`:

```bash
curl -L --fail  https://github.com/jaegertracing/jaeger/releases/download/v1.40.0/jaeger-1.40.0-windows-amd64.tar.gz -o jaeger-1.40.0-windows-amd64.tar.gz 
tar -xzf jaeger-1.40.0-windows-amd64.tar.gz
./jaeger-1.40.0-windows-amd64/jaeger-all-in-one
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

Build the observability server:

```bash
docker build -f ./server/Dockerfile -t grpc-productinfo-observability-server:latest .
```

Run the observability server (Windows git bash):

```bash
winpty docker run -it grpc-productinfo-observability-server:latest
```

Run client and server:

```bash
docker build -f ./server/Dockerfile -t grpc-productinfo-observability-server:latest .
docker build -f ./client/Dockerfile -t grpc-productinfo-observability-client:latest .

docker network create product-net
winpty docker run -it --network=product-net --hostname=productinfo -p 50051:50051 grpc-productinfo-observability-server:latest
winpty docker run -it --network=product-net --hostname=client --env PRODUCT_INFO_SERVER=productinfo grpc-productinfo-observability-client:latest
```

