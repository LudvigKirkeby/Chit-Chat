package main

import (
	"context"
	"google.golang.org/grpc"
	proto "handin-3/grpc"
	"log"
	"net"
	"time"
)

type Client struct {
	ID     int          // might not be needed
	Stream proto.MessageService_ChatServer
}

type System struct {
	proto.UnimplementedMessageServiceServer
	mutex      sync.Mutex         // lock to avoid weird data problems
	clients map[int]*Client
	nextID int
}

func main() {
    server := &system{clients: make(map[int]*Client), nextID: 1}
	server.start_server()
}

func (s *System) broadcast(msg *proto.Message) {
    s.mutex.lock()
    defer s.mutex.Unlock()
    log.Printf("Broadcasting message to clients")

    for id, client := range s.client {
        err := client.Stream.Send(&proto.Message{
        LamportTime: time.Now().UnixNano(),
        Text: text,
        })
    if err != nil {
            log.Printf("Error sending to Client %d: %v", id, err)
        }
    }
    s.mu.unlock()
}

func (s *system) SendMessage(ctx context.Context, in *proto.Message) (*proto.Empty, error){
    //log that the message vas received in the server via RPC
    log.Printf("Received message via sendMessage RPC")
    //call broadcast function to share message with everyone
    s.broadcast(in)
}

// Should be "Broadcast"; should output messages from users to everyone
func (s *system) GetMessages(ctx context.Context, in *proto.Empty) (*proto.Message, error) {
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

// joins a client, gives id, and strats goroutine
func (s *system) Join(stream proto.MessageService_ChatServer) error {
	s.mu.Lock() // lock so data doesnt go weird
	clientID := s.nextID++ // starts at 1, remember ++ is post-increment
	client := &Client{ID: clientID, Stream: stream} // instantiate a new client
	s.clients[clientID] = client // map client to its ID
	s.mu.Unlock()

    // join msg
	s.broadcast(clientID, fmt.Sprintf("Participant %d joined Chit Chat", clientID))

	// Dedicated goroutine for this client
	go func() { // go func runs a concurrent function
		for {
			msg, err := stream.Recv() // waits for client msg
			if err != nil {
			    s.removeClient(clientID)
			    s.broadcast(clientID, "Participant left chat")
			    // should remove the client and broadcast that they disconnected
				return
			}
            // broadcast msg
			s.broadcast(clientID, msg.GetRelevantIdentifiers())
		}
	}()

	// Keep RPC open until client disconnects
	select {}
}

// deletes client from list
func (s *Server) removeClient(id int) {
	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()
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


	err = grpcserver.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}

//func (s *system) GetTime(ctx context.Context, in *proto.Empty) (*proto.Message, error) {
//	return &proto.Message{Time: []string{time.Now().String()}}, nil
//}

