package main

import (
	"log"
	"strings"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

const (
	exampleScheme      = "custom"
	exampleServiceName = "pb.example.grpc.io"
)

var addrs = []string{"localhost:50051", "localhost:50052"}

type exampleResolverBuilder struct{}
type exampleResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (*exampleResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := exampleResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			exampleServiceName: addrs, // "pb.example.grpc.io": "localhost:50051", "localhost:50052"
		},
	}
	r.start()
	return r, nil
}

func (*exampleResolverBuilder) Scheme() string {
	return exampleScheme
}

func (r *exampleResolver) start() {
	url := strings.TrimLeft(r.target.URL.Path, "/")

	addresses := r.addrsStore[url]
	addrs := make([]resolver.Address, len(addresses))
	leader := true
	for i, s := range addresses {
		addrs[i] = resolver.Address{
			Addr:       s,
			Attributes: attributes.New("is_leader", leader),
		}
		leader = false
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (exampleResolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (exampleResolver) Close() {}

func init() {
	log.Println("Register exampleResolver")
	resolver.Register(&exampleResolverBuilder{})
}
