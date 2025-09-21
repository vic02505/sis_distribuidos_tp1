package utils

import (
	"log"
	"strconv"
	"sync"
	"time"
)

type task struct {
	taskId         uint8
	taskType       string
	taskStatus     string
	assignedWorker *string
	timeStamp      time.Time
}

type SharedResources struct {
	mutex         sync.Mutex
	mapsToDo      uint8
	reducesToDo   uint8
	reducerAmount uint8
	tasksMap      map[string]task
}

func CreateInitialSharedResources(fileSplits []string, reducerAmount uint8) *SharedResources {

	taskMap := make(map[string]task)

	i := 1
	for _, fileSplit := range fileSplits {
		taskMap[fileSplit] = task{taskId: uint8(i), taskStatus: NotAssigned, assignedWorker: nil,
			timeStamp: time.Now(), taskType: Map}
		i += 1
	}

	reducerNumber := 1
	for range reducerAmount {
		fileName := "mr-x-" + strconv.Itoa(reducerNumber)
		taskMap[fileName] = task{taskId: uint8(reducerNumber), taskStatus: NotAssigned, assignedWorker: nil,
			timeStamp: time.Now(), taskType: Reduce}
		reducerNumber += 1
	}

	return &SharedResources{
		tasksMap: taskMap,
	}
}

func (sr *SharedResources) GetAndAssignAvailableWork(workerUuid string) (*string, *uint8) {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	var availableWork *string

	if sr.reducesToDo == 0 {
		log.Printf("Thre is no more work to do!!")
	}

	if sr.mapsToDo > 0 {
		availableWork = sr.getFirstAvailableMappingTask()
	} else if (sr.mapsToDo == 0) && (sr.reducesToDo > 0) {
		availableWork = sr.getFirstAvailableReduceTask()
	} else {
		return nil, nil
	}

	if availableWork == nil {
		return nil, nil
	}

	sr.assignTask(*availableWork, workerUuid)
	taskId := sr.tasksMap[*availableWork].taskId

	return availableWork, &taskId
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
	task.taskStatus = Finished
	sr.tasksMap[workToMark] = task
}
