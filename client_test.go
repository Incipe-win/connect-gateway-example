package main

import (
	"bookstore_enhance/proto"
	"bookstore_enhance/proto/protoconnect"
	"context"
	"log"
	"net/http"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func createClient() protoconnect.BookStoreClient {
	client := protoconnect.NewBookStoreClient(http.DefaultClient, "http://localhost:8888")
	return client
}

func TestListShelves(t *testing.T) {
	client := createClient()

	req := &emptypb.Empty{}

	resp, err := client.ListShelves(context.Background(), connect.NewRequest(req))
	if err != nil {
		t.Fatalf("ListShelves failed: %v", err)
	}
	if resp == nil || len(resp.Msg.Shelves) == 0 {
		t.Fatal("Expected non-empty shelves response")
	}

	log.Printf("ListShelves response: %v", resp.Msg.Shelves)
}

func TestCreateShelf(t *testing.T) {
	client := createClient()

	req := &proto.CreateShelfRequest{
		Shelf: &proto.Shelf{
			Theme: "科幻",
		},
	}

	resp, err := client.CreateShelf(context.Background(), connect.NewRequest(req))
	if err != nil {
		t.Fatalf("CreateShelf failed: %v", err)
	}

	log.Printf("CreateShelf response: %v", resp.Msg)
}

func TestListBooks(t *testing.T) {
	client := createClient()

	req := &proto.ListBooksRequest{
		Shelf: 5,
	}

	resp, err := client.ListBooks(context.Background(), connect.NewRequest(req))
	if err != nil {
		t.Fatalf("ListBooks failed: %v", err)
	}
	log.Printf("ListBooks response: %v", resp.Msg.Books)
}
