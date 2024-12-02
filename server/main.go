package main

import (
	"bytes"
	"context"
	"io"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"net"
	// "strings"

	// "github.com/google/uuid"
	pb "github.com/jerryhong21/todo-grpc/proto"
	"google.golang.org/grpc"
	// "google.golang.org/protobuf/types/known/emptypb"
)

// this is where i implement the functions
/**
1. CreateTodo
2. GetTodo
3. PostTodo
4. DeleteTodo
5. ListTodo

*/

// This is the standard gRPC method signature in Go
// func (s *server) MethodName(ctx context.Context, req *RequestType) (*ResponseType, error)

type server struct {
	pb.UnimplementedTodoServiceServer
	todos map[string]*pb.Todo // maps todo Ids to todo
}

func NewServer() *server {
	return &server {
		todos: make(map[string]*pb.Todo),
	}
}

type CreateTodoPayload struct {
	TaskID      string `json:"task_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CreateTodo
// Returns a pb.Todo object
func (s *server) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.Todo, error) {

	// send a request to SC API to create todo
	SC_ACTIONS_URL := "https://api.safetyculture.io/tasks/v1/actions"

	// Create 
	payloadData := CreateTodoPayload{
		TaskID: req.GetId(),
		Title: req.GetTitle(),
		Description: req.GetDescription(),
	}

	
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		fmt.Printf("Failed to encode payload: %v", err)
		return nil, fmt.Errorf("internal server error")
	}

	payload := bytes.NewReader(payloadBytes)

	// Make the HTTP req to SC platform
	httpReq, err := http.NewRequest("POST", SC_ACTIONS_URL, payload)
	if err != nil {
		fmt.Printf("Failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("internal server error")
	}
	// Add relevant details to the header
	httpReq.Header.Add("accept", "application/json")
	httpReq.Header.Add("content-type", "application/json")
	API_KEY := os.Getenv("SC_API_KEY")
	httpReq.Header.Add("authorization", "Bearer " + API_KEY)

	// Retrieve response
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		fmt.Printf("CreateTodo: error requesting SC API: %v", err)
    	return nil, fmt.Errorf("failed to send request to SafetyCulture API: %w", err)
	}
	
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to process response from SafetyCulture API")
	}

	fmt.Println("Successfully retrieved API - here is the response from SafetyCulture API")
	fmt.Println(string(body))
	
	// return the pb.Todo 
	responseTodo := &pb.Todo{
		Id: req.GetId(),
		Title: req.GetTitle(),
		Description: req.GetDescription(),
		Completed: false,
	}
	
	// Populate the server data
	s.todos[req.GetId()] = responseTodo

	return responseTodo, nil
}


// Main server
func main() {

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTodoServiceServer(grpcServer, NewServer())
	log.Println("gRPC server is running on port :50051")
	if err := grpcServer.Serve(lis); err != nil {
    	log.Fatalf("Failed to serve: %v", err)
	}

}