package communications

import (
	"context"
	"log"
	"tp1/coordinator/internal/utils"
	pb "tp1/protocol/messages"
)

type communicationHandler struct {
	pb.UnimplementedServerServer
	sharedResources *utils.SharedResources
}

func (c *communicationHandler) AskForWork(ctx context.Context, req *pb.ImFree) (*pb.AskForWorkResponse, error) {
	log.Printf("Someone asked for work")
	if c.sharedResources.IsThereAvailableWork() {
		log.Printf("Worker<%s> wants job", req.WorkerUuid)
		assignedTask := c.sharedResources.AssignMappingWork(req.WorkerUuid)
		resp := utils.BuildAskForWorkResponse(assignedTask, 1, "Map", 1)
		log.Printf("Assigned job to Worker<%s>", req.WorkerUuid)
		return resp, nil
	} else {
		log.Printf("There's no work avalaible")
		return &pb.AskForWorkResponse{Response: "No work"}, nil
	}
}

func (c *communicationHandler) MarkWorkAsFinished(ctx context.Context, req *pb.IFinished) (*pb.IFinishedResponse, error) {
	log.Printf("A worker finished a job")
	c.sharedResources.MarkWorkAsFinished(req.WorkerUuid, "")
	return &pb.IFinishedResponse{Response: "OK"}, nil
}
