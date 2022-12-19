FROM golang AS build

ENV BUILD_DIR=/build
WORKDIR ${BUILD_DIR}
ADD . ${BUILD_DIR}
RUN go get -d ./...
RUN go install ./...
RUN CGO_ENABLED=0 go build -o /bin/grpc-productinfo-server ./server_prometheus

FROM scratch
COPY --from=build /bin/grpc-productinfo-server /bin/grpc-productinfo-server
ENTRYPOINT ["/bin/grpc-productinfo-server"]
EXPOSE 50051
