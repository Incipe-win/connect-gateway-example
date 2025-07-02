package main

import (
	"fmt"

	"google.golang.org/grpc/resolver"
)

const (
	myScheme   = "hchao"
	myEndpoint = "resolver.incipe.com"
)

var addrs = []string{"127.0.0.1:8888", "127.0.0.1:8972", "127.0.0.1:8973"}

type hchaoResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *hchaoResolver) ResolveNow(o resolver.ResolveNowOptions) {
	addrStrs := r.addrsStore[r.target.Endpoint()]
	addrList := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrList[i] = resolver.Address{Addr: s}
	}
	r.cc.UpdateState(resolver.State{
		Addresses: addrList,
	})
}

func (r *hchaoResolver) Close() {}

type hchaoResolverBuilder struct {
}

func (*hchaoResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &hchaoResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			myEndpoint: addrs,
		},
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

func (*hchaoResolverBuilder) Scheme() string {
	return myScheme
}

func init() {
	fmt.Printf("Registering custom resolver with scheme: %s\n", myScheme)
	resolver.Register(&hchaoResolverBuilder{})
}
