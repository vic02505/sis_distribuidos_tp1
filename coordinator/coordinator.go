package main

import (
	"context"
	"log"
	"net"
	pb "tp1/protocol/messages"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedServerServer
	assignedWorkers []string
}

func (s *Server) AskForWork(ctx context.Context, req *pb.ImFree) (*pb.AskForWorkResponse, error) {
	if len(s.assignedWorkers) < 3 {
		s.assignedWorkers = append(s.assignedWorkers, req.Content)
		return &pb.AskForWorkResponse{Response: "Work"}, nil
	} else {
		return &pb.AskForWorkResponse{Response: "No work"}, nil
	}
}

func main() {
	socketPath := "/tmp/mr-socket.sock"

	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Error al escuchar: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterServerServer(grpcServer, &Server{})

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error al servir: %v", err)
	}

}
