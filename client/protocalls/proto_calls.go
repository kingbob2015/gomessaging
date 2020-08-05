package protocalls

import (
	"context"

	"github.com/kingbob2015/gomessaging/proto/messagingpb"
)

//RegisterAsClient takes as input a display name to register a user with and returns their id
func RegisterAsClient(ctx context.Context, c messagingpb.MessagingServiceClient, name string) (string, error) {
	res, err := c.RegisterAsClient(
		ctx,
		&messagingpb.RegisterAsClientRequest{
			DisplayName: name,
		},
	)
	if err != nil {
		return "", err
	}
	return res.GetUserId(), nil
}

//OpenReceiveChannel takes a user id already registered with the messaging service and returns a stream to receive messages on
func OpenReceiveChannel(ctx context.Context, c messagingpb.MessagingServiceClient, id string) (messagingpb.MessagingService_OpenReceiveChannelClient, error) {
	recStream, err := c.OpenReceiveChannel(
		ctx,
		&messagingpb.OpenReceiveChannelRequest{
			UserId: id,
		},
	)
	if err != nil {
		return nil, err
	}
	return recStream, nil
}
