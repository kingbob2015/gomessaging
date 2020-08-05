package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kingbob2015/gomessaging/proto/messagingpb"
	"github.com/kingbob2015/gomessaging/server/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//TODO: Should move all this server code into a new file
type server struct {
	messagingpb.UnimplementedMessagingServiceServer
	//Going to keep the clientList here as a pointer for now and as a named field instead of embedded type
	//This could go either way but going to keep as pointer in case we want to share clients among multiple servers
	//Going to keep as named field to keep constructor style in client list utilities
	clientList *utils.ClientList
}

func newServer() *server {
	return &server{clientList: utils.NewClientList()}
}

func (s *server) serverValidCheck() error {
	if s.clientList == nil {
		return errors.New("The client list of the server is uninitialized")
	}
	return nil
}

func (s *server) RegisterAsClient(ctx context.Context, req *messagingpb.RegisterAsClientRequest) (*messagingpb.RegisterAsClientResponse, error) {
	err := s.serverValidCheck()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("An error occurred during server validation: %v", err),
		)
	}

	name := req.GetDisplayName()
	id, err := s.clientList.AddNewClient(name)
	//improvement here: we should have different error types and return diff codes for each type
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("An error occurred during adding of client: %v", err),
		)
	}

	return &messagingpb.RegisterAsClientResponse{
		UserId: id,
	}, nil
}

func (s *server) OpenReceiveChannel(req *messagingpb.OpenReceiveChannelRequest, stream messagingpb.MessagingService_OpenReceiveChannelServer) error {
	id := req.GetUserId()
	err := s.clientList.SetClientConn(id, &stream)
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("There was an error opening the receiving connection: %v", err),
		)
	}
	return nil
}

func (s *server) GetClientList(ctx context.Context, req *messagingpb.GetClientListRequest) (*messagingpb.GetClientListResponse, error) {
	list, err := s.clientList.GetList()
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error getting the list of users: %v", err),
		)
	}
	return &messagingpb.GetClientListResponse{
		ClientList: list,
	}, nil
}

func (s *server) SendMessage(ctx context.Context, req *messagingpb.SendMessageRequest) (*messagingpb.SendMessageResponse, error) {
	recID, err := s.clientList.GetUserID(req.GetReceiverDisplayName())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error getting receiver user id: %v", err),
		)
	}

	stream, err := s.clientList.GetClientConn(recID)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error getting receiver user connection: %v", err),
		)
	}
	dispName, err := s.clientList.GetUserName(req.GetUserId())
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Error getting sender user details: %v", err),
		)
	}
	(*stream).Send(&messagingpb.OpenReceiveChannelResponse{
		SenderDisplayName: dispName,
		Message:           req.GetMessage(),
	})
	return &messagingpb.SendMessageResponse{}, nil
}

func main() {
	// if we crash the go code, we get the file name and line number in error message when we use log
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Messaging Service Started")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer lis.Close()

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	defer s.Stop()

	serv := newServer()
	messagingpb.RegisterMessagingServiceServer(s, serv)

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
	fmt.Println("End of Program")
}
