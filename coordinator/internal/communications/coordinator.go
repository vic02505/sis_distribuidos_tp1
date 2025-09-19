package communications

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"tp1/coordinator/internal/utils"
	pb "tp1/protocol/messages"
)

type Coordinator struct {
	communicationHandler *communicationHandler
	sharedResources      *utils.SharedResources
	mappersAmount        uint8
	reducersAmount       uint8
}

func NewCoordinator(fileSplits []string, reducers uint8) *Coordinator {

	sharedResources := utils.CreateInitialSharedResources(fileSplits)

	return &Coordinator{
		communicationHandler: &communicationHandler{sharedResources: sharedResources},
		sharedResources:      sharedResources,
		mappersAmount:        uint8(len(fileSplits)),
		reducersAmount:       reducers,
	}
}

func (c *Coordinator) StartCoordinator() {
	socketPath := "/tmp/mr-socket.sock"

	lis, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Socket listening error: %v", err)
	}

	grpcServer := grpc.NewServer()

	pb.RegisterServerServer(grpcServer, c.communicationHandler)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Error al servir: %v", err)
	}

}
