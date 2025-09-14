package main

import (
	"fmt"
	"context"
	"log"

	pb "tp1/protocol/messages"
	"google.golang.org/grpc"
)

func main() {
	socketPath := "/tmp/mr-socket.sock"

	conn, err := grpc.Dial( "unix://" + socketPath, grpc.WithInsecure())

	client := pb.NewServerClient(conn)

	resp, err := client.AskForWork(context.Background(), &pb.ImFree{Content: "Worker 1"})
	if err != nil {
		log.Fatalf("Error al solicitar trabajo: %v", err)
	}

	fmt.Println(resp.Response)
}
