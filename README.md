# Todo-grpc

This application is created to help me learn Go, gRPC and protobufs.

**Todo-grpc** is a lightweight Todo-list application built in Go, utilizing gRPC and Protocol Buffers for efficient and scalable client-server communication. This application offers comprehensive CRUD functionalities, enabling users to manage their tasks seamlessly. 


## Technology Stack

- **Go:** The primary language for both server and client implementations, known for its performance and concurrency support.
- **gRPC:** A high-performance RPC framework facilitating seamless and type-safe communication between client and server.
- **Protocol Buffers:** An efficient interface definition language used to define service contracts and serialize structured data.
- **SafetyCulture API Integration:** Integrates with the SafetyCulture API to provide advanced features such as bulk deletion of todo items.

## Features

- **Create Todo:** Add new todo items with unique identifiers.
- **Read Todo:** Retrieve existing todo items by their IDs.
- **Update Todo:** Modify details of existing todos.
- **Delete Todo:** Remove individual or multiple todos efficiently.
- **Bulk Deletion:** Utilize SafetyCulture API for deleting multiple todos in a single operation.
