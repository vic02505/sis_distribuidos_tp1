package utils

import pb "tp1/protocol/messages"

func BuildAskForWorkResponse(assignedTask string, workerId int32, workType string, reducerAmount uint8) *pb.AskForWorkResponse {
	return &pb.AskForWorkResponse{FilePath: assignedTask, WorkType: workType,
		WorkerId: workerId, ReducerNumber: int32(reducerAmount)}
}
