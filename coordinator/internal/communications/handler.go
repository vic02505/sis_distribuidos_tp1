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
	shutdownChan    chan bool
}

func (c *communicationHandler) AskForWork(ctx context.Context, req *pb.ImFree) (*pb.AskForWorkResponse, error) {
	log.Printf("Someone asked for work")

	workToDo := c.sharedResources.GetAndAssignAvailableWork(req.WorkerUuid)

	if workToDo != nil {
		log.Printf("Worker<%s> wants job", req.WorkerUuid)
		resp := utils.BuildAskForWorkResponse(workToDo.WorkName, int32(workToDo.Task.TaskId), workToDo.Task.TaskType, workToDo.ReducerAmount)
		log.Printf("Assigned job to Worker<%s>", req.WorkerUuid)
		return resp, nil

	} else {
		log.Printf("There's no work avalaible")
		return &pb.AskForWorkResponse{WorkType: "Work finished"}, nil
	}
}

func (c *communicationHandler) MarkWorkAsFinished(ctx context.Context, req *pb.IFinished) (*pb.IFinishedResponse, error) {
	log.Printf("A worker finished a job")
	c.sharedResources.MarkWorkAsFinished(req.WorkFinished, req.WorkType)

	if c.sharedResources.IsAllWorkCompleted() {
		select {
		case c.shutdownChan <- true:
		default:
		}
	}

	return &pb.IFinishedResponse{Response: "OK"}, nil
}
