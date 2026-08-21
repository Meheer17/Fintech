package main

import (
	"context"
	"log"
	"time"

	pb "github.com/RevenueIQ/revenueiq_dev_kit/proto/mongo_service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func main() {
	log.Println("Connecting to gRPC server at localhost:50051...")
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	userClient := pb.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	log.Println("\n==========================================")
	log.Println("=== TESTING RICH USER SERVICE FLOWS ===")
	log.Println("==========================================")

	// Step 1: Create User
	log.Println("--- 1. Testing CreateUser ---")
	createReq := &pb.CreateUserRequest{
		FirstName:   "Alice",
		LastName:    "Wonderland",
		Email:       "alice@wonderland.com",
		Age:         24,
		Password:    "super_secret_pwd",
		Role:        "Admin",
		PhoneNumber: "+1234567890",
	}
	createResp, err := userClient.CreateUser(ctx, createReq)
	if err != nil {
		log.Fatalf("CreateUser failed: %v", err)
	}
	user := createResp.GetUser()
	userID := user.GetId()
	log.Printf("Created User: ID=%s, Name=%s, Email=%s, Role=%s, Phone=%s, CreatedAt=%d",
		userID, user.GetName(), user.GetEmail(), user.GetRole(), user.GetPhoneNumber(), user.GetCreatedAt())

	// Wait briefly to allow the background cache writing goroutine to execute
	time.Sleep(200 * time.Millisecond)

	// Step 2: Get User (Should hit Cache)
	log.Println("\n--- 2. Getting User (Should hit Cache) ---")
	getResp, err := userClient.GetUser(ctx, &pb.GetUserRequest{Id: userID})
	if err != nil {
		log.Fatalf("GetUser failed: %v", err)
	}
	log.Printf("Retrieved User: ID=%s, Name=%s, Email=%s (Cached hit works!)",
		getResp.GetUser().GetId(), getResp.GetUser().GetName(), getResp.GetUser().GetEmail())

	// Step 3: Get User By Email
	log.Println("\n--- 3. Getting User By Email ---")
	getByEmailResp, err := userClient.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{Email: "alice@wonderland.com"})
	if err != nil {
		log.Fatalf("GetUserByEmail failed: %v", err)
	}
	log.Printf("Retrieved User By Email: ID=%s, Name=%s, Active=%t",
		getByEmailResp.GetUser().GetId(), getByEmailResp.GetUser().GetName(), getByEmailResp.GetUser().GetActive())

	// Step 4: Query Users (Pagination & Search with Dynamic Filters)
	log.Println("\n--- 4. Querying Users (Search & Pagination with Dynamic Filters) ---")
	queryResp, err := userClient.QueryUsers(ctx, &pb.QueryUsersRequest{
		Page:   1,
		Limit:  5,
		Search: "Alice",
		Active: true,
		Filters: map[string]string{
			"role": "Admin",
		},
	})
	if err != nil {
		log.Fatalf("QueryUsers failed: %v", err)
	}
	log.Printf("Query returned %d user(s) on Page %d. Total count = %d, Total pages = %d",
		len(queryResp.GetUsers()), queryResp.GetCurrentPage(), queryResp.GetTotalCount(), queryResp.GetTotalPages())
	for _, u := range queryResp.GetUsers() {
		log.Printf("  -> User: Name=%s, Email=%s, Active=%t", u.GetName(), u.GetEmail(), u.GetActive())
	}

	// Step 5: Update User
	log.Println("\n--- 5. Updating User (Invalidates Cache) ---")
	updateResp, err := userClient.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:            userID,
		FirstName:     proto.String("Alice"),
		LastName:      proto.String("Wonderland-Smith"),
		Email:         proto.String("alice.smith@wonderland.com"),
		Age:           proto.Int32(25),
		Password:      proto.String("updated_secret_pwd"),
		Role:          proto.String("SuperAdmin"),
		PhoneNumber:   proto.String("+1987654321"),
		EmailVerified: proto.Bool(true),
		PhoneVerified: proto.Bool(true),
		Active:        proto.Bool(true),
	})
	if err != nil {
		log.Fatalf("UpdateUser failed: %v", err)
	}
	log.Printf("Updated User: Name=%s, Email=%s, EmailVerified=%t, PhoneVerified=%t",
		updateResp.GetUser().GetName(), updateResp.GetUser().GetEmail(), updateResp.GetUser().GetEmailVerified(), updateResp.GetUser().GetPhoneVerified())

	// Step 6: Delete User
	log.Println("\n--- 6. Deleting User ---")
	deleteResp, err := userClient.DeleteUser(ctx, &pb.DeleteUserRequest{Id: userID})
	if err != nil {
		log.Fatalf("DeleteUser failed: %v", err)
	}
	log.Printf("Delete result: Success=%t", deleteResp.GetSuccess())

	// Verify Deletion
	log.Println("\n--- 7. Verifying Deletion ---")
	_, err = userClient.GetUser(ctx, &pb.GetUserRequest{Id: userID})
	if err != nil {
		log.Printf("Successfully confirmed user deletion: %v", err)
	} else {
		log.Fatalf("Error: Deleted user was still found!")
	}
}
