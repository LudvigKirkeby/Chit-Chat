package main

import (
	"bufio"
	"context"
	"fmt"
	proto "handin-3/grpc" // adjust to your generated package path
	"log"
	"os"
	"strings"

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
		log.Fatalf("Error opening stream for client %v, err: %v", client, err)
	}

	recvCh := make(chan *proto.Message)
	sendCh := make(chan string)

	// receiver goroutine: prints any broadcasts
	go func() {
		// Local lamport
		for {
			msg, err := stream.Recv()
			if err != nil {
				log.Println("Server has probably been closed. Further error message:", err)
				return
			}
			recvCh <- msg
			//fmt.Printf("[%v] %s\n", Time.Format("2006-01-02 15:04:05"), msg.GetText())
		}
	}()

	// main goroutine: read stdin and send
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Type messages and press Enter to send.")
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			sendCh <- line
		}
		// close send side and exit
		close(sendCh)
	}()

	// essentially "main" for the goroutines
	// local lamport set to 0
	var localClock = int64(0)

	for {
		// open says if channel still open
		if sendCh == nil && recvCh == nil {
			break
		}
		select {
		case line, open := <-sendCh:
			if !open {
				_ = stream.CloseSend()
				sendCh = nil
				continue
			}

			// increment clock cause we are sending something
			localClock++

			if len(line) > 128 {
				fmt.Printf("Your message was not sent. Reason: Too large")
				continue
			}

			// define msg to send, use "line" from ch
			out := &proto.Message{
				LamportClock: localClock,
				Text:         line,
			}

			if err := stream.Send(out); err != nil {
				log.Println("error on sending: ", err)
			}

		case msg, open := <-recvCh:
			if !open {
				_ = stream.CloseSend()
				recvCh = nil
				continue
			}
			localClock = max(localClock, msg.LamportClock) + 1
			fmt.Printf("[Logical Time %d] %s\n", localClock, msg.GetText())
		}
	}
}
