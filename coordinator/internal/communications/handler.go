package communications

import (
	"context"
	"github.com/google/uuid"
	"tp1/coordinator/internal/utils"
	pb "tp1/protocol/messages"
)

type communicationHandler struct {
	pb.UnimplementedServerServer
	sharedResources *utils.SharedResources
}

func (c *communicationHandler) AskForWork(ctx context.Context, req *pb.ImFree) (*pb.AskForWorkResponse, error) {
	if c.sharedResources.IsThereAvailableWork() {
		assignedTask := c.sharedResources.AssignMappingWork(uuid.New().String())
		return &pb.AskForWorkResponse{Response: assignedTask}, nil
	} else {
		return &pb.AskForWorkResponse{Response: "No work"}, nil
	}
}
