package utils

import (
	"log"
	"strconv"
	"sync"
	"time"
)

type Task struct {
	TaskId         uint8
	TaskType       string
	TaskStatus     string
	AssignedWorker *string
	TimeStamp      *time.Time
}

type SharedResources struct {
	mutex         sync.Mutex
	mapsToDo      uint8
	reducesToDo   uint8
	reducerAmount uint8
	tasksMap      map[string]Task
}

type WorkToDo struct {
	WorkName      string
	Task          Task
	ReducerAmount uint8
}

func CreateInitialSharedResources(fileSplits []string, reducerAmount uint8) *SharedResources {

	taskMap := make(map[string]Task)

	i := 1
	for _, fileSplit := range fileSplits {
		taskMap[fileSplit] = Task{TaskId: uint8(i), TaskStatus: NotAssigned, AssignedWorker: nil,
			TimeStamp: nil, TaskType: Map}
		i += 1
	}

	reducerNumber := 1
	for range reducerAmount {
		fileName := "mr-x-" + strconv.Itoa(reducerNumber)
		taskMap[fileName] = Task{TaskId: uint8(reducerNumber), TaskStatus: NotAssigned, AssignedWorker: nil,
			TimeStamp: nil, TaskType: Reduce}
		reducerNumber += 1
	}

	return &SharedResources{
		tasksMap:      taskMap,
		mapsToDo:      uint8(len(fileSplits)),
		reducesToDo:   reducerAmount,
		reducerAmount: reducerAmount,
	}
}

func (sr *SharedResources) GetAndAssignAvailableWork(workerUuid string) *WorkToDo {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	var workName *string
	var workToDo *Task

	if sr.mapsToDo > 0 {
		workName, workToDo = sr.getFirstAvailableMappingTask()
	} else if (sr.mapsToDo == 0) && (sr.reducesToDo > 0) {
		workName, workToDo = sr.getFirstAvailableReduceTask()
	} else {
		log.Printf("There is no more work to do!!")
		return nil
	}

	if workName == nil || workToDo == nil {
		return nil
	}

	sr.assignTask(*workName, workerUuid)

	return &WorkToDo{WorkName: *workName, Task: *workToDo, ReducerAmount: 3}
}

func (sr *SharedResources) MarkWorkAsFinished(workToMark string, workType string) {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	if workType == Map && sr.mapsToDo > 0 {
		sr.mapsToDo -= 1
	}

	if workType == Reduce && sr.reducesToDo > 0 {
		sr.reducesToDo -= 1
	}

	task := sr.tasksMap[workToMark]
	task.TaskStatus = Finished
	sr.tasksMap[workToMark] = task
}

func (sr *SharedResources) IsAllWorkCompleted() bool {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	return sr.mapsToDo == 0 && sr.reducesToDo == 0
}
