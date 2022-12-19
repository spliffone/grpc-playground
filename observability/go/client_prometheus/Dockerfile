FROM golang AS build

ENV BUILD_DIR=/build
WORKDIR ${BUILD_DIR}
ADD . ${BUILD_DIR}
RUN go get -d ./...
RUN go install ./...
RUN CGO_ENABLED=0 go build -o /bin/grpc-productinfo-client ./client_opencensus

FROM scratch
COPY --from=build /bin/grpc-productinfo-client /bin/grpc-productinfo-client
ENTRYPOINT ["/bin/grpc-productinfo-client"]
EXPOSE 50051
