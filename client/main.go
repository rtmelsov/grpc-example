package main

import (
	"context"
	pb "demo/proto"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"log"
)

func TestUsers(c pb.UsersClient) {

	users := []*pb.User{
		{Name: "a", Email: "a@example.com", Sex: pb.User_MALE},
		{Name: "b", Email: "b@example.com", Sex: pb.User_FEMALE},
		{Name: "c", Email: "c@example.com", Sex: pb.User_MALE},
		{Name: "d", Email: "d@example.com", Sex: pb.User_UNSPECIFIED},
		{Name: "e", Email: "e@example.com", Sex: pb.User_FEMALE},
		{Name: "f", Email: "f@example.com", Sex: pb.User_FEMALE},
	}

	for _, user := range users {
		md := metadata.New(map[string]string{"token": "12345"})
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		resp, err := c.AddUser(ctx, &pb.AddUserRequest{
			User: user,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Response: %s", resp.String())
	}

	resp, err := c.DelUser(context.Background(), &pb.DelUserRequest{
		Email: "a@example.com",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %s", resp.String())

	for _, userEmail := range []string{"b@example.com", "a@example.com"} {
		resp, err := c.GetUser(context.Background(), &pb.GetUserRequest{
			Email: userEmail,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(resp.String())
	}

	emails, err := c.ListUsers(context.Background(), &pb.ListUsersRequest{
		Offset: 0,
		Limit:  100,
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(emails.Count, emails.Emails)
}

func main() {
	conn, err := grpc.NewClient(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewUsersClient(conn)

	TestUsers(c)
}
