package main

import (
	"context"
	"fmt"
	"net"

	"github.com/svwielga4/grpc/server/echo"
	"google.golang.org/grpc"
)

type EchoServer struct{}

// Echo is a function that responds to a request
func (e *EchoServer) Echo(ctx context.Context, req *echo.EchoRequest) (*echo.EchoResponse, error) {
	return &echo.EchoResponse{
		Response: "Echo: " + req.Message,
	}, nil
}

func main() {
	lst, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	srv := &EchoServer{}
	echo.RegisterEchoServerServer(s, srv)

	fmt.Println("Now serving at port 8080")
	err = s.Serve(lst)
	if err != nil {
		panic(err)
	}

}
