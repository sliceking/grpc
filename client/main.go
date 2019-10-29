package main

import (
	"context"
	"fmt"
	"os"

	"github.com/svwielga4/grpc/chat"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("must have a url as first arg, and username as second arg")
		return
	}

	ctx := context.Background()

	conn, err := grpc.Dial(os.Args[1], grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := chat.NewChatClient(conn)
	stream, err := c.Chat(ctx)
	if err != nil {
		panic(err)
	}

	waitc := make(chan struct{})
	go func(){
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			} else if err != nil{
				panic(err)
			}
			fmt.Println(msg.User + ": " + msg.Message)
		}
	}

}
