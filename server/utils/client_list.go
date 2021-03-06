package utils

import (
	"errors"

	"github.com/google/uuid"
	"github.com/kingbob2015/gomessaging/proto/messagingpb"
)

//ClientList is a type for holding a client list. Should always be created via NewClientList()
type ClientList struct {
	list []*client
}

type client struct {
	id         string
	name       string
	connection Connection
}

//Connection is a type for holding a user's connection, needed to be a struct to keep it alive
type Connection struct {
	Conn  *messagingpb.MessagingService_OpenReceiveChannelServer
	Close chan bool
}

//NewClientList is a constructor for the ClientList type
func NewClientList() *ClientList {
	s := make([]*client, 0, 0)
	return &ClientList{list: s}
}

//GetList returns the list of display names of clients currently registered
func (c *ClientList) GetList() ([]string, error) {
	if c.list == nil {
		return []string{}, errors.New("ClientList list is not initialized")
	}
	ret := make([]string, len(c.list))
	for i, cl := range c.list {
		ret[i] = cl.name
	}
	return ret, nil
}

//GetUserID takes a user's display name and returns their uuid
func (c *ClientList) GetUserID(name string) (string, error) {
	if c.list == nil {
		return "", errors.New("ClientList list is not initialized")
	}
	for _, cl := range c.list {
		if cl.name == name {
			return cl.id, nil
		}
	}
	return "", errors.New("User does not exist")
}

//GetUserName takes a user's id and returns their uuid
func (c *ClientList) GetUserName(id string) (string, error) {
	if c.list == nil {
		return "", errors.New("ClientList list is not initialized")
	}
	for _, cl := range c.list {
		if cl.id == id {
			return cl.name, nil
		}
	}
	return "", errors.New("User does not exist")
}

func (c *ClientList) userNameExists(name string) bool {
	for _, cl := range c.list {
		if cl.name == name {
			return true
		}
	}
	return false
}

//AddNewClient generates a uuid for a client and adds them to the client list and returns the uuid as a string
func (c *ClientList) AddNewClient(name string) (string, error) {
	if c.list == nil {
		return "", errors.New("ClientList list is not initialized")
	}
	if c.userNameExists(name) {
		return "", errors.New("User name is in use")
	}
	id := uuid.New().String()
	c.list = append(c.list, &client{id: id, name: name})
	return id, nil
}

//SetClientConn takes a stream that is used to send messages to a client and adds it to a user's record
func (c ClientList) SetClientConn(id string, stream *messagingpb.MessagingService_OpenReceiveChannelServer) (Connection, error) {
	if c.list == nil {
		return Connection{}, errors.New("ClientList list is not initialized")
	}
	for i, cl := range c.list {
		if cl.id == id {
			conn := Connection{
				Conn:  stream,
				Close: make(chan bool),
			}
			c.list[i].connection = conn
			return conn, nil
		}
	}
	return Connection{}, errors.New("User does not exist")
}

//GetClientConn takes a user id and returns the stream for receiving messages for that id's user
func (c ClientList) GetClientConn(id string) (*messagingpb.MessagingService_OpenReceiveChannelServer, error) {
	if c.list == nil {
		return nil, errors.New("ClientList list is not initialized")
	}
	for i, cl := range c.list {
		if cl.id == id {
			stream := c.list[i].connection.Conn
			return stream, nil
		}
	}
	return nil, errors.New("User does not exist")
}
