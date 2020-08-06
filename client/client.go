package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"

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
	//Register the user
	id := register(&c)
	fmt.Printf("Registered with id: %v\n", id)

	//Open a receive channel
	fmt.Println("Opening a channel to receive messages...")
	stream, err := openRecvChannel(&c, id)
	if err != nil {
		log.Fatalf("Failed to open receiving stream: %v", err)
	}

	//Thread off to a receiving go routine
	go receiveMessages(stream)

	//Provide user options to send messages and get client list
	go userInteraction(&c, id)

	//Make a channel that takes in when ctrl+c is hit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	//Block until a signal is received
	<-ch
	fmt.Println("End of Program")
}

func register(c *messagingpb.MessagingServiceClient) string {
	fmt.Println("Please set a display name: ")
	in := bufio.NewReader(os.Stdin)

	userName, err := in.ReadString('\n')
	if err != nil {
		log.Fatalf("Error reading input: %v\n", err)
	}
	userName = trimNewLineFromReadString(userName)
	//Register as a user
	id, err := protocalls.RegisterAsClient(context.Background(), c, userName)
	for err != nil {
		log.Printf("Error from user creation: %v\n", err)
		fmt.Println("Please try again to set a display name: ")
		userName, err = in.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading input: %v\n", err)
		}
		userName = trimNewLineFromReadString(userName)
		id, err = protocalls.RegisterAsClient(context.Background(), c, userName)
	}
	return id
}

func openRecvChannel(c *messagingpb.MessagingServiceClient, id string) (*messagingpb.MessagingService_OpenReceiveChannelClient, error) {
	recStream, err := protocalls.OpenReceiveChannel(context.Background(), c, id)
	if err != nil {
		return nil, err
	}
	return recStream, nil
}

func receiveMessages(stream *messagingpb.MessagingService_OpenReceiveChannelClient) {
	for {
		msg, err := (*stream).Recv()
		if err == io.EOF {
			//we've reached the end of the stream (the stream was closed)
			log.Printf("Server has cut off connection: %v", err)
			break
		}
		if err != nil {
			log.Printf("Error while reading stream: %v", err)
		}
		sender := msg.GetSenderDisplayName()
		message := msg.GetMessage()
		fmt.Printf("User %v has sent the following message: \n%v\n", sender, message)
	}
}

func userInteraction(c *messagingpb.MessagingServiceClient, id string) {
	in := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Enter 1 to get a list of other users or 2 to send a message: ")

		opt, err := in.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v\n", err)
			continue
		}
		opt = trimNewLineFromReadString(opt)
		switch opt {
		case "1":
			list, err := protocalls.GetClientList(context.Background(), c)
			if err != nil {
				log.Printf("Failed to get client list: %v\n", err)
				continue
			}
			for _, client := range list {
				fmt.Println(client)
			}
		case "2":
			list, err := protocalls.GetClientList(context.Background(), c)
			if err != nil {
				log.Printf("Failed to get client list: %v\n", err)
				continue
			}
			fmt.Println("Enter which user you would like to send a message to: ")
			user, err := in.ReadString('\n')
			if err != nil {
				log.Printf("Error reading input: %v\n", err)
				continue
			}
			user = trimNewLineFromReadString(user)
			//Check to see if we have the user in the list
			i := 0
			for _, client := range list {
				if client == user {
					break
				}
				i++
			}
			if i == len(list) {
				log.Printf("%v is not a valid user from the current user list\n", err)
				continue
			}
			fmt.Printf("Enter your message to send to %v\n", user)
			msg, err := in.ReadString('\n')
			if err != nil {
				log.Printf("Error reading input: %v\n", err)
				continue
			}
			msg = trimNewLineFromReadString(msg)

			err = protocalls.SendMessage(context.Background(), c, id, user, msg)
			if err != nil {
				log.Printf("Failed to send message: %v\n", err)
				continue
			}
			fmt.Println("Successfully sent the message!")
		default:
			fmt.Printf("Invalid option: %v\n", opt)
			continue
		}
	}
}

func trimNewLineFromReadString(in string) string {
	if runtime.GOOS == "windows" {
		return strings.TrimRight(in, "\r\n")
	} else {
		return strings.TrimRight(in, "\n")
	}
}
