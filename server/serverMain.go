package main

import (
	"context"
	"google.golang.org/grpc"
	proto "handin-3/grpc"
	"log"
	"net"
	"time"
)

func main() {
	start_server()
}

type server struct {
	proto.UnimplementedTimeServiceServer
}

func (s *server) GetTime(ctx context.Context, in *proto.Empty) (*proto.Time, error) {
	return &proto.Time{Time: time.Now().String()}, nil
}

func start_server() {
	grpcserver := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	proto.RegisterTimeServiceServer(grpcserver, &server{})

	err = grpcserver.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}
