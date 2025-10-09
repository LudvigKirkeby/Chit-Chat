package main

import (
	"context"
	proto "handin-3/grpc"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect")
	}

	client := proto.NewTimeServiceClient(conn)

	log.SetFlags(0) // Disable inbuilt timestamp
	time, err := client.GetTime(context.Background(), &proto.Empty{})
	if err != nil {
		log.Fatalf("issue with time code %v", err)
	}
	log.Println("The time is: " + time.Time)
}
