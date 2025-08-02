package main

import (
	"context"
	pb "demo/proto"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"sort"
	"sync"
	"time"
)

type UsersServer struct {
	pb.UnimplementedUsersServer
	users sync.Map
}

func (s *UsersServer) AddUser(ctx context.Context, in *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	var response pb.AddUserResponse
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("token")
		if len(values) > 0 {
			token := values[0]
			fmt.Println("token", token)
		}
	}

	if _, ok := s.users.Load(in.User.Email); ok {
		return nil, status.Errorf(codes.Aborted, "User with email %v doesn't exists", in.User.Email)
	} else {
		s.users.Store(in.User.Email, in.User)
	}
	return &response, nil
}

func (s *UsersServer) ListUsers(ctx context.Context, in *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {

	var list []string

	s.users.Range(func(key, _ any) bool {
		list = append(list, key.(string))
		return true
	})

	sort.Strings(list)

	offset := int(in.Offset)
	end := int(in.Offset + in.Limit)
	if end > len(list) {
		end = len(list)
	}

	if offset >= end {
		offset = 0
		end = 0
	}

	response := pb.ListUsersResponse{
		Count:  int32(len(list)),
		Emails: list[offset:end],
	}

	return &response, nil
}

func ServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	fmt.Println("Incoming request...")
	resp, err := handler(ctx, req)

	fmt.Println("Finished!", info.FullMethod, "Duration", time.Since(start))

	if err != nil {
		st, _ := status.FromError(err)
		fmt.Println("Error: ", st.Message())
	}

	return resp, nil
}

func (s *UsersServer) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	var response pb.GetUserResponse

	if user, ok := s.users.Load(in.Email); ok {
		response.User = user.(*pb.User)
	} else {
		return nil, status.Errorf(codes.NotFound, "User with email %v doesn't exists", in.Email)
	}
	return &response, nil
}
func (s *UsersServer) DelUser(ctx context.Context, in *pb.DelUserRequest) (*pb.DelUserResponse, error) {
	var response pb.DelUserResponse

	if _, ok := s.users.LoadAndDelete(in.Email); !ok {
		return nil, status.Errorf(codes.NotFound, "User with email %v doesn't exists", in.Email)
	}
	return &response, nil
}
func main() {
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}

	// middlewares for grpc = interceptors
	s := grpc.NewServer(
		grpc.UnaryInterceptor(ServerInterceptor),
	)
	pb.RegisterUsersServer(s, &UsersServer{})

	if err = s.Serve(listen); err != nil {
		log.Fatal(err)
	}
}
