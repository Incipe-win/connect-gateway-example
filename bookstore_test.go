package main

import (
	"bookstore_enhance/proto"
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test_server_ListBooks(t *testing.T) {
	db, _ := NewDB("bookstore.db")
	s := &server{
		bs: &bookstore{
			db: db,
		},
	}

	req := &proto.ListBooksRequest{
		Shelf: 5,
	}
	res, err := s.ListBooks(context.Background(), connect.NewRequest(req))
	if err != nil {
		t.Fatalf("ListBooks failed: %v", err)
	}
	t.Logf("next page token: %s", res.Msg.NextPageToken)
	for i, book := range res.Msg.Books {
		t.Logf("%d: %#v\n", i+1, book)
	}
}

func TestScheme(t *testing.T) {
	conn, err := grpc.NewClient("hchao:///resolver.incipe.com",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewBookStoreClient(conn)
	req := &proto.ListBooksRequest{
		Shelf: 5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := c.ListBooks(ctx, req)
	if err != nil {
		t.Fatalf("ListBooks failed: %v", err)
	}
	t.Logf("books: %v", res.Books)
}
