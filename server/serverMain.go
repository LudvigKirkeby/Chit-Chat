package main

import (
	"fmt"
	proto "handin-3/grpc"
	"log"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"
)

type Client struct {
	ID     int // might not be needed
	Stream proto.MessageService_JoinServer
}

type system struct {
	proto.UnimplementedMessageServiceServer
	mutex      sync.Mutex // lock to avoid weird data problems
	clients    map[int]*Client
	nextID     int
	localClock int64
}

var serverLogger *log.Logger

func main() {
	port := 8080
	f, err := os.OpenFile("chitchat.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		serverLogger.Fatalf("failed to open log file: %v", err)
	}
	// ensure it's closed on exit
	defer f.Close()

	// create a server-specific logger (doesn't change global logger)
	flags := log.Ldate | log.Ltime | log.Lmicroseconds
	serverLogger = log.New(f, "[SERVER] ", flags)

	// use serverLogger.Printf instead of log.Printf
	serverLogger.Printf("[STARTUP] Startup port=%d", port)

	server := &system{clients: make(map[int]*Client), nextID: 1, localClock: 0}
	server.start_server()
}

func (s *system) broadcast(msg string) {

	for id, client := range s.clients {
		s.mutex.Lock()
		s.localClock++
		s.mutex.Unlock()
		serverLogger.Printf("[SEND] '%v' to client %v at logical time %d", msg, id, s.localClock)
		err := client.Stream.Send(&proto.Message{
			LamportClock: s.localClock,
			Text:         msg,
		})
		if err != nil {
			serverLogger.Printf("Error sending to Client %d: %v", id, err)
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
	serverLogger.Printf("[CONNECT] Client %d joined Chit Chat at logical time %v", clientID, s.localClock)

	// join msg
	s.broadcast(fmt.Sprintf("Client %d joined Chit Chat at logical time %v", clientID, s.localClock))

	// Dedicated goroutine for this client
	go func() { // go func runs a concurrent function
		for {
			msg, err := stream.Recv() // waits for client msg
			if err != nil {
				s.removeClient(clientID)
				serverLogger.Printf("[DISCONNECT] Client %d left Chit Chat at logical time %v", clientID, s.localClock)
				s.broadcast(fmt.Sprintf("Client %d left Chit Chat at logical time %v", clientID, s.localClock))
				// should remove the client and broadcast that they disconnected
				return
			}
			s.mutex.Lock()
			s.localClock = max(s.localClock, msg.LamportClock) + 1
			serverLogger.Printf("[RECEIVE] '%v' from client %d at Logical time %d", msg.Text, clientID, msg.GetLamportClock())
			s.mutex.Unlock()
			// broadcast msg, this might be wrong
			s.broadcast(fmt.Sprintf("Client %d said: %v", clientID, msg.GetText()))
		}
	}()

	// Keep RPC open until client disconnects
	select {}
}

// deletes client from list
func (s *system) removeClient(id int) {
	s.mutex.Lock()
	s.localClock++
	delete(s.clients, id)
	s.mutex.Unlock()
}

func (s *system) start_server() {
	// Listens for requests on port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		serverLogger.Fatal(err)
	}

	// Makes new GRPC server
	grpc_server := grpc.NewServer()

	// Registers the grpc server with the System struct
	proto.RegisterMessageServiceServer(grpc_server, s)

	err = grpc_server.Serve(listener)
	if err != nil {
		serverLogger.Fatal(err)
	}

	serverLogger.Printf("grpc server %v started with listener %v at TCP address 8080", grpc_server, listener)
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
