package main

import (
	"fmt"
	"google.golang.org/grpc"
)

type server struct {
	UnimplementedOrderServiceServer
}

func main() {
	s := grpc.NewServer()
	RegisterOrderServiceServer(s, &server{})
	fmt.Println("Dogfood Order service listening on :50051")
}
