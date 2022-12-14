# Resolver

Basic client name resolver example. Instead of using Load-Balancer proxy we use a client side fail over.

Client Features:

- Custom name resolver

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
