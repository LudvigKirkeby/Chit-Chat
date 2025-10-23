package main

import (
	"context"
	proto "handin-3/grpc"
	"log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
    log.SetFlags(0) // Disable inbuilt timestamp
    sendMessage()
}

func sendMessage() {
 conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("could not connect")
	}

    client := proto.NewMessageServiceClient(conn)

    stream, err := client.Join(context.Background(), &proto.Empty{})
        	if err != nil {
        		log.Printf("Client %d: Join error: %v", id, err)
        		return
        	}

    scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
    		line := scanner.Text()
    		if strings.TrimSpace(line) == "" {
    			continue
    		}
    		if err := stream.Send(&proto.Message{Text: line}); err != nil {
    			log.Println("send error:", err)
    			break
    }
}


