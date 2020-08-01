package main

import (
	"log"

	"google.golang.org/grpc"
)

func main() {
	opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect %v", err)
	}

	defer cc.Close()
}
