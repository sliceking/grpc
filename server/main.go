package main

import (
	"io"

	"github.com/svwielga4/grpc/chat"
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
