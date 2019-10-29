package main

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/svwielga4/grpc/chat"
	"google.golang.org/grpc"
)

// Connection struct used to communicate through
type Connection struct {
	conn chat.Chat_ChatServer
	send chan *chat.ChatMessage
	quit chan struct{}
}

// NewConnection creates a new connection
func NewConnection(conn chat.Chat_ChatServer) *Connection {
	c := &Connection{
		conn: conn,
		send: make(chan *chat.ChatMessage),
		quit: make(chan struct{}),
	}
	go c.start()
	return c
}

// Close will close the connection
func (c *Connection) Close() error {
	close(c.quit)
	close(c.send)
	return nil
}

// Send will put messages on the channel to communicate
func (c *Connection) Send(msg *chat.ChatMessage) {
	defer func() {
		// Ignore any errors about sending on a closed channel
		recover()
	}()
	c.send <- msg
}

func (c *Connection) start() {
	running := true
	for running {
		select {
		case msg := <-c.send:
			c.conn.Send(msg)
		case <-c.quit:
			running = false
		}
	}
}

// GetMessages will get messages off the connection and send the message through a broadcast channel
func (c *Connection) GetMessages(broadcast chan<- *chat.ChatMessage) error {
	for {
		msg, err := c.conn.Recv()
		if err == io.EOF {
			c.Close()
			return nil
		} else if err != nil {
			c.Close()
			return err
		}
		go func(msg *chat.ChatMessage) {
			select {
			case broadcast <- msg:
			case <-c.quit:
			}
		}(msg)
	}
}

type ChatServer struct {
	broadcast   chan *chat.ChatMessage
	quit        chan struct{}
	connections []*Connection
	connLock    sync.Mutex
}

func NewChatServer() *ChatServer {
	srv := &ChatServer{
		broadcast: make(chan *chat.ChatMessage),
		quit:      make(chan struct{}),
	}
	go srv.start()
	return srv
}

// Close is a global closer for all channels
func (c *ChatServer) Close() error {
	close(c.quit)
	return nil
}

func (c *ChatServer) start() {
	running := true
	for running {
		select {
		case msg := <-c.broadcast:
			c.connLock.Lock()
			for _, v := range c.connections {
				go v.Send(msg)
			}
			c.connLock.Unlock()
		}
	}
}

func (c *ChatServer) Chat(stream chat.Chat_ChatServer) error {
	conn := NewConnection(stream)

	c.connLock.Lock()
	c.connections = append(c.connections, conn)
	c.connLock.Unlock()

	err := conn.GetMessages(c.broadcast)

	c.connLock.Lock()
	for i, v := range c.connections {
		if v == conn {
			c.connections = append(c.connections[:i], c.connections[i+1:]...)
		}
	}
	c.connLock.Unlock()

	return err
}

func main() {
	lst, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	srv := NewChatServer()
	chat.RegisterChatServer(s, srv)

	fmt.Println("Now serving at port 8080")
	err = s.Serve(lst)
	if err != nil {
		panic(err)
	}
}
