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
	server := &database{messages: []Event{}}
	server.start_server()
}

type database struct {
	proto.UnimplementedTimeServiceServer
	messages []Event
}

type Event struct {
	Timestamp           string
	ComponentName       string
	EventType           string
	RelevantIdentifiers string
}

func (s *database) GetMessages(ctx context.Context, in *proto.Empty) (*proto.Message, error) {
	var timestamps, componentNames, eventTypes, relevantIds []string
	for _, e := range s.messages {
		timestamps = append(timestamps, e.Timestamp)
		componentNames = append(componentNames, e.ComponentName)
		eventTypes = append(eventTypes, e.EventType)
		relevantIds = append(relevantIds, e.RelevantIdentifiers)
	}
	return &proto.Message{
		Time:                timestamps,
		ComponentName:       componentNames,
		EventType:           eventTypes,
		RelevantIdentifiers: relevantIds,
	}, nil
}

//func (s *database) SendMessage(ctx context.Context, in *proto.Message) (*proto.Empty, error) {
//	s.timestamp = append(s.timestamp, m.timestamp)
//}

func (s *database) GetTime(ctx context.Context, in *proto.Empty) (*proto.Message, error) {
	return &proto.Message{Time: time.Now().String()}, nil
}

func (s *database) start_server() {
	grpcserver := grpc.NewServer()
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	proto.RegisterTimeServiceServer(grpcserver, s)

	err = grpcserver.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}
