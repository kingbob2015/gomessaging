package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/kingbob2015/gomessaging/proto/messagingpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	messagingpb.UnimplementedMessagingServiceServer
}

// func (*server) RegisterAsClient(context.Context, *RegisterAsClientRequest) (*RegisterAsClientResponse, error) {

// }

// func (*server) OpenReceiveChannel(*ReceiveChannelRequest, MessagingService_OpenReceiveChannelServer) error {

// }

// func (*server) GetClientList(context.Context, *GetClientListRequest) (*GetClientListResponse, error) {

// }

// func (*server) SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error) {

// }

func main() {
	// if we crash the go code, we get the file name and line number in error message when we use log
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Messaging Service Started")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	messagingpb.RegisterMessagingServiceServer(s, &server{})

	reflection.Register(s)

	go func() {
		fmt.Println("Starting Server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve %v", err)
		}
	}()

	//Make a channel that takes in when ctrl+c is hit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	//Block until a signal is received
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("End of Program")
}
