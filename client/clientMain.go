package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	proto "handin-3/grpc" // adjust to your generated package path

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := proto.NewMessageServiceClient(conn)
	stream, err := client.Join(context.Background())
	if err != nil {
		log.Fatalf("open Join stream: %v", err)
	}

	// receiver goroutine: prints any broadcasts
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Println("receive error (server closed?):", err)
				return
			}
			fmt.Printf("[broadcast] %s\n", msg.GetText())
		}
	}()

	// main goroutine: read stdin and send
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type messages and press Enter to send.")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if err := stream.Send(&proto.Message{Text: line}); err != nil {
			log.Println("send error:", err)
			break
		}
	}
	// close send side and exit
	_ = stream.CloseSend()
}
