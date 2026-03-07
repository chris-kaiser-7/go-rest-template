package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"

	userpb "github.com/chris-a-kaiser-7/go-rest-template/cmd/grpc/proto/user"
)

func main() {
	// Connect to the server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	// Create client
	client := userpb.NewUserManagerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// Add a user
	addResp, err := client.RegisterUser(ctx, &userpb.AddUserRequest{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Password: "bestpassword1234",
	})
	if err != nil {
		log.Fatalf("AddUser failed: %v", err)
	}

	log.Printf("User added successfully! ID: %s", addResp.Id)
	// Get the user
	// getResp, err := client.GetUser(ctx, &userpb.GetUserRequest{
	// 	Id: addResp.Id,
	// })
	// if err != nil {
	// 	log.Fatalf("GetUser failed: %v", err)
	// }
	// if getResp.Found {
	// 	user := getResp.User
	// 	log.Printf("Retrieved user: %s (%s), Age: %d", user.Name, user.Email, user.Age)
	// } else {
	// 	log.Println("User not found")
	// }
}
