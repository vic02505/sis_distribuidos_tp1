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

	availableWork, workId := c.sharedResources.GetAndAssignAvailableWork(req.WorkerUuid)

	if availableWork != nil {
		log.Printf("Worker<%s> wants job", req.WorkerUuid)
		resp := utils.BuildAskForWorkResponse(*availableWork, int32(*workId), "Map", 3)
		log.Printf("Assigned job to Worker<%s>", req.WorkerUuid)
		return resp, nil

	} else {
		log.Printf("There's no work avalaible")
		return &pb.AskForWorkResponse{WorkType: "Work finished"}, nil
	}
}

func (c *communicationHandler) MarkWorkAsFinished(ctx context.Context, req *pb.IFinished) (*pb.IFinishedResponse, error) {
	log.Printf("A worker finished a job")
	c.sharedResources.MarkWorkAsFinished(req.WorkFinished)
	return &pb.IFinishedResponse{Response: "OK"}, nil
}
