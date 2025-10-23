package main

import (
	"fmt"
	"google.golang.org/grpc"
	proto "handin-3/grpc"
	"log"
	"net"
	"sync"
	"time"
)

type Client struct {
	ID     int // might not be needed
	Stream proto.MessageService_JoinServer
}

type system struct {
	proto.UnimplementedMessageServiceServer
	mutex   sync.Mutex // lock to avoid weird data problems
	clients map[int]*Client
	nextID  int
}

func main() {
	server := &system{clients: make(map[int]*Client), nextID: 1}
	server.start_server()
}

func (s *system) broadcast(msg string) {
	log.Printf("Broadcasting message to clients")

	for id, client := range s.clients {
		err := client.Stream.Send(&proto.Message{
			LamportTime: time.Now().UnixNano(),
			Text:        msg,
		})
		if err != nil {
			log.Printf("Error sending to Client %d: %v", id, err)
		}
	}
}

// joins a client, gives id, and strats goroutine
func (s *system) Join(stream proto.MessageService_JoinServer) error {
	s.mutex.Lock()       // lock so data doesnt go weird
	clientID := s.nextID // starts at 1
	s.nextID++
	client := &Client{ID: clientID, Stream: stream} // instantiate a new client
	s.clients[clientID] = client                    // map client to its ID
	s.mutex.Unlock()

	// join msg
	s.broadcast(fmt.Sprintf("Participant %d joined Chit Chat", clientID))

	// Dedicated goroutine for this client
	go func() { // go func runs a concurrent function
		for {
			msg, err := stream.Recv() // waits for client msg
			if err != nil {
				s.removeClient(clientID)
				s.broadcast(fmt.Sprintf("Participant %d left chat", clientID))
				// should remove the client and broadcast that they disconnected
				return
			}
			// broadcast msg
			s.broadcast(fmt.Sprintf("Participant %d said: %v", clientID, msg.GetText()))
		}
	}()

	// Keep RPC open until client disconnects
	select {}
}

// deletes client from list
func (s *system) removeClient(id int) {
	s.mutex.Lock()
	delete(s.clients, id)
	s.mutex.Unlock()
}

func (s *system) start_server() {
	// Listens for requests on port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	// Makes new GRPC server
	grpc_server := grpc.NewServer()

	// Registers the grpc server with the System struct
	proto.RegisterMessageServiceServer(grpc_server, s)

	err = grpc_server.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}

//func (s *system) GetTime(ctx context.Context, in *proto.Empty) (*proto.Message, error) {
//	return &proto.Message{Time: []string{time.Now().String()}}, nil
//}

/*
func (s *system) SendMessage(ctx context.Context, in *proto.Message) (*proto.Empty, error) {
	//log that the message vas received in the server via RPC
	log.Printf("Received message via sendMessage RPC")
	//call broadcast function to share message with everyone
	msg, err := context.Background(), in.Text
	if err != nil {
		return
	}
	s.broadcast(msg)
}

*/
