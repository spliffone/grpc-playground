# Basics

Basic client server example.

Server Features:

- Simple logging interceptor for stream and unary requests
- Enrich response interceptor

Client Features:

- Request timeout
- Pass additional headers to each RPC request
- Enhanced RPC error status parsing

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
