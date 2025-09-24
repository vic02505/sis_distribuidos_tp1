package utils

import (
	"log"
	"time"
)

func (sr *SharedResources) getFirstAvailableMappingTask() (*string, *Task) {

	var taskName string
	var taskToDo Task

	for fileSplit, task := range sr.tasksMap {
		if (task.TaskType == Map) && (task.TaskStatus == NotAssigned) {
			taskName = fileSplit
			taskToDo = task
			break
		} else if (task.TaskType == Map) && (task.TimeStamp != nil) && (task.TaskStatus == Assigned) {
			if time.Since(*task.TimeStamp) > 10*time.Second {
				log.Printf("A worker died!")
				taskName = fileSplit
				taskToDo = task
				break
			}
		}
	}

	return &taskName, &taskToDo
}

func (sr *SharedResources) getFirstAvailableReduceTask() (*string, *Task) {

	var taskName string
	var taskToDo Task

	for fileSplit, task := range sr.tasksMap {
		if (task.TaskType == Reduce) && (task.TaskStatus == NotAssigned) {
			taskName = fileSplit
			taskToDo = task
			break
		} else if (task.TaskType == Reduce) && (task.TimeStamp != nil) && (task.TaskStatus == Assigned) {
			if time.Since(*task.TimeStamp) > 10*time.Second {
				log.Printf("A worker died!")
				taskName = fileSplit
				taskToDo = task
				break
			}
		}
	}

	return &taskName, &taskToDo
}

func (sr *SharedResources) assignTask(workToAssign, workerUuid string) {

	currentTime := time.Now()

	task := sr.tasksMap[workToAssign]
	task.TaskStatus = Assigned
	task.TimeStamp = &currentTime
	task.AssignedWorker = &workerUuid
	sr.tasksMap[workToAssign] = task

}
