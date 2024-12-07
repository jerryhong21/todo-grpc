package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"github.com/joho/godotenv"

	pb "github.com/jerryhong21/todo-grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// this is where i implement the functions
/**
1. CreateTodo - DONE
2. GetTodo
3. PostTodo
4. BulkDeleteTodo - DONE
5. ListTodo
*/

// This is the standard gRPC method signature in Go
// func (s *server) MethodName(ctx context.Context, req *RequestType) (*ResponseType, error)

type server struct {
	pb.UnimplementedTodoServiceServer
	todos map[string]*pb.Todo // maps todo Ids to todo
}

func NewServer() *server {
	return &server{
		todos: make(map[string]*pb.Todo),
	}
}

type CreateTodoPayload struct {
	TaskID      string `json:"task_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GetTodoPayload struct {
	TaskID string `json:"task_id"`
}

type BulkDeleteTodoPayload struct {
	TaskIDs []string `json:"ids"`
}

// // Standard struct for SafetyCulture API error response
// type ScErrorResponse struct {
// 	Code int `json:"code"`
// 	Message string `json:"message"`
// 	Details []any `json:"details"`
// }

// CreateTodo
// Returns a pb.Todo object
// context.Context is a type interaface (inherently a pointer) and therefore does not need a pointer

func (s *server) CreateTodo(ctx context.Context, req *pb.CreateTodoRequest) (*pb.Todo, error) {

	// send a request to SC API to create todo
	SC_ACTIONS_URL := "https://api.safetyculture.io/tasks/v1/actions"

	// Create
	payloadData := CreateTodoPayload{
		TaskID:      req.GetId(),
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
	}

	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		fmt.Printf("Failed to encode payload: %v", err)
		return nil, fmt.Errorf("internal server error")
	}

	payload := bytes.NewReader(payloadBytes)

	// Make the HTTP req to SC platform
	httpReq, err := http.NewRequestWithContext(ctx, "POST", SC_ACTIONS_URL, payload)
	if err != nil {
		fmt.Printf("Failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("internal server error")
	}
	// Add relevant details to the header
	httpReq.Header.Add("accept", "application/json")
	httpReq.Header.Add("content-type", "application/json")
	API_KEY := os.Getenv("SC_API_KEY")
	httpReq.Header.Add("authorization", "Bearer "+API_KEY)

	// Retrieve response
	res, _:= http.DefaultClient.Do(httpReq)
	resBody, err := handleResponse(res)
	if resBody == nil && err != nil {
		fmt.Printf("The API returned with an error: %v", err)
		return nil, err
	}

	fmt.Println("Successfully retrieved API - here is the response from SafetyCulture API")
	fmt.Println(string(resBody))

	// return the pb.Todo
	responseTodo := &pb.Todo{
		Id:          req.GetId(),
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Completed:   false,
	}

	// Populate the server data
	s.todos[req.GetId()] = responseTodo

	return responseTodo, nil
}

// Delete a todo item using bulk delete API
// Returns nothing
func (s *server) BulkDeleteTodo(ctx context.Context, req *pb.BulkDeleteTodoRequest) (*emptypb.Empty, error) {

	// concatenates delete action prefix and actionId
	ids := req.GetIds()
	SC_DELETE_TODO_URL := "https://api.safetyculture.io/tasks/v1/actions/delete"

	// Create a payload struct, JSONIFY it
	payloadData := BulkDeleteTodoPayload{
		TaskIDs: ids,
	}
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		fmt.Printf("Failed to encode payload: %v", err)
		return nil, fmt.Errorf("internal server error")
	}
	payload := bytes.NewReader(payloadBytes)

	// The below works for single deletion payload, the above works for bulk deletion
	// payload := strings.NewReader(fmt.Sprintf(`{"ids":["%s"]}`, id))

	deleteReq, _ := http.NewRequest("POST", SC_DELETE_TODO_URL, payload)

	deleteReq.Header.Add("accept", "application/json")
	deleteReq.Header.Add("content-type", "application/json")
	deleteReq.Header.Add("authorization", "Bearer "+os.Getenv("SC_API_KEY"))

	res, _ := http.DefaultClient.Do(deleteReq)

	body, err := handleResponse(res)
	if body == nil && err != nil {
		fmt.Println("The API returned with an error: %w", err)
		return nil, err
	}

	// if body is not nil, but empty - this signifies correct
	if len(body) == 0 {
		// remove the todo from our body
		for _, id := range ids {
			titleRemoved := s.todos[id]
			delete(s.todos, id)
			fmt.Printf("Successfully deleted %v\n", titleRemoved)
		}
		return &emptypb.Empty{}, nil
	}

	// TODO: Experiment and see if ids contain partially valid ids, then does the API remove the valid ones and return error?
	// If so, then we need to update the behaviour of our function such that the valid IDs are removed

	return nil, fmt.Errorf("error: Response from SC API:\n %v", string(body))
}

func (s *server) GetTodo(ctx context.Context, req *pb.GetTodoRequest) (*pb.Todo, error) {
	fmt.Println("HELLOOOOO")
	id := req.GetId()
	SC_GET_TODO_URL := "https://api.safetyculture.io/tasks/v1/actions/" + id

	fmt.Printf("url is %v\n", SC_GET_TODO_URL)

	getReq, _ := http.NewRequest("GET", SC_GET_TODO_URL, nil)
	getReq.Header.Add("accept", "application/json")

	res, _ := http.DefaultClient.Do(getReq)
	resBody, err := handleResponse(res)
	if resBody == nil && err != nil {
		fmt.Println("The API returned with an error: %w", err)
		return nil, err
	}

	// else, return the response body
	return s.todos[id], nil
}

// func (s *server) UpdateTodo(ctx context.Context, req *pb.UpdateTodoRequest) (*pb.Todo, error) {
// 	id := req.GetId()

// }

// function to handle response
// only checks for whether there are any errors
func handleResponse(res *http.Response) ([]byte, error) {
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v", err)
		return nil, fmt.Errorf("failed to process response from SafetyCulture API")
	}
	fmt.Println(string(body))

	// decode the response into a map of keys of type string, which maps to values of ANY kind
	var result map[string]any
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// if the field "code" exists, then error
	_, exists := result["code"]
	if !exists {
		fmt.Println("Normal response")
		return body, nil
	}

	// Else, print the error
	return nil, fmt.Errorf("%s", string(body))
	/*

		IF WE WANT OT HANDLE THE ERROR REPSONSE:

		errRes := ScErrorResponse{}

		// decode the json string contained in the "mesage field of the results map"
		// breakdown:
		// result["message"] is of type "any" - which means we need to make a type assertion: .(string)
		// []byte(...) converts the string value into a byte slice since json.Unmarshal requires a byte slice
		// finally, json.Unmarshal decodes JSON data, into a go strct
		err = json.Unmarshal([]byte(result["message"].(string)), &errRes)
	*/
}

// Main server
func main() {

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	err = godotenv.Load("../.env")
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

	grpcServer := grpc.NewServer()
	pb.RegisterTodoServiceServer(grpcServer, NewServer())
	log.Println("gRPC server is running on port :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
