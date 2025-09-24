package communications

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"tp1/coordinator/internal/utils"
	pb "tp1/protocol/messages"
)

type Coordinator struct {
	communicationHandler *communicationHandler
	sharedResources      *utils.SharedResources
	mappersAmount        uint8
	reducersAmount       uint8
	shutdownChan         chan bool
}

func NewCoordinator(fileSplits []string, reducersAmount uint8) *Coordinator {

	sharedResources := utils.CreateInitialSharedResources(fileSplits, reducersAmount)
	shutdownChan := make(chan bool, 1)

	return &Coordinator{
		communicationHandler: &communicationHandler{sharedResources: sharedResources, shutdownChan: shutdownChan},
		sharedResources:      sharedResources,
		mappersAmount:        uint8(len(fileSplits)),
		reducersAmount:       reducersAmount,
		shutdownChan:         shutdownChan,
	}
}

func (c *Coordinator) StartCoordinator() {
	socketPath := "/tmp/mr-socket.sock"

	os.Remove(socketPath)

	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Socket listening error: %v", err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()

	pb.RegisterServerServer(grpcServer, c.communicationHandler)

	log.Printf("Coordinator listening...")

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server stopped: %v", err)
		}
	}()

	<-c.shutdownChan
	log.Printf("All work completed. Shutting down...")

	grpcServer.GracefulStop()
	os.Remove(socketPath)
}
