package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kingbob2015/gomessaging/client/protocalls"
	"github.com/kingbob2015/gomessaging/proto/messagingpb"
	"google.golang.org/grpc"
)

func main() {
	opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect %v", err)
	}

	defer cc.Close()

	c := messagingpb.NewMessagingServiceClient(cc)

	var in string
	fmt.Println("Please set a display name: ")
	fmt.Scanln(&in)
	//Register as a user
	id, err := protocalls.RegisterAsClient(context.Background(), c, in)
	for err != nil {
		log.Fatalf("Error from user creation: %v\n", err)
		fmt.Println("Please try again to set a display name: ")
		fmt.Scanln(&in)
		id, err = protocalls.RegisterAsClient(context.Background(), c, in)
	}

	//Open a receive channel

}
