package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	pb "github.com/jerryhong21/todo-grpc/proto"
	"google.golang.org/grpc"
	// "google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldn't connect to server: %v", err)
	}

	defer conn.Close()

	client := pb.NewTodoServiceClient(conn)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\nTodo CLI")
		fmt.Println("1. Create Todo")
		fmt.Println("2. Get Todo")
		fmt.Println("3. Update Todo")
		fmt.Println("4. Delete Todo")
		fmt.Println("5. List Todos")
		fmt.Println("6. Exit")
		fmt.Print("Choose an option: ")

		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		switch option {
		case "1":
			createTodo(client, reader)
		// case "2":
		//     getTodo(client, reader)
		// case "3":
		//     updateTodo(client, reader)
		case "4":
			bulkDeleteTodo(client, reader)
		// case "5":
		//     listTodos(client)
		case "6":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid option")
		}
	}

}

// Sends the server a request
// pb.TodoServiceClient is a type interface.
// the client parameter is NOT a pointer, because interfaces in Go are inherently reference types
// so no need, reader, is a different story
func createTodo(client pb.TodoServiceClient, reader *bufio.Reader) {

	fmt.Print("Enter TODO ID: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	// validate ID in UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		fmt.Printf("Inputted %v is not a valid UUID\n", id)
		return
	}

	fmt.Print("Enter title: ")
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)

	fmt.Print("Enter description: ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	// to prevent client hangs, we time it out for one second
	// craetes a context with a timeout
	// this means any operation using this context must complete within 1 second
	// so we pass this into our gRPC calls to the server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// "defer" schedules the cancel() call to run when the current function completes
	defer cancel()

	// passing ctx into this function ensures this function lasts 1 second at most
	res, err := client.CreateTodo(ctx, &pb.CreateTodoRequest{
		Id:          id,
		Title:       title,
		Description: description,
	})

	if err != nil {
		fmt.Printf("Error creating todo: %v", err)
		return
	}

	// prints out the response from server
	jsonData, _ := json.MarshalIndent(res, "", "  ")
	fmt.Printf("Created Todo:\n %s", jsonData)
}

// TODO: Implement bulk deletion functionality
// Make the CLI ask for multiple ids, isntead of just one
func bulkDeleteTodo(client pb.TodoServiceClient, reader *bufio.Reader) {

	fmt.Print("Enter Todo ID to delete: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	// checking for valid ID
	_, err := uuid.Parse(id)
	if err != nil {
		fmt.Printf("Inputted %v is not a valid UUID\n", id)
		return
	}

	// add the singular id to the string of ids
	ids := []string{}
	ids = append(ids, id)

	// create context and wait
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// "defer" schedules the cancel() call to run when the current function completes
	defer cancel()

	_, err = client.BulkDeleteTodo(ctx, &pb.BulkDeleteTodoRequest{
		Ids: ids,
	})

	if err != nil {
		fmt.Printf("Error deleting todo: %v", err)
		return
	}

	fmt.Printf("Successfully deleted todo item:\n %s", id)
}
