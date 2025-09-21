package utils

func (sr *SharedResources) getFirstAvailableMappingTask() *string {

	var availableTask *string = nil

	for fileSplit, task := range sr.tasksMap {
		if task.taskStatus == NotAssigned && task.taskType == Map {
			aux := fileSplit
			availableTask = &aux
			break
		}
	}

	return availableTask
}

func (sr *SharedResources) getFirstAvailableReduceTask() *string {

	var availableTask *string = nil

	for fileSplit, task := range sr.tasksMap {
		if task.taskStatus == NotAssigned && task.taskType == Reduce {
			aux := fileSplit
			availableTask = &aux
			break
		}
	}

	return availableTask
}

func (sr *SharedResources) assignTask(workToAssign, workerUuid string) {
	task := sr.tasksMap[workToAssign]
	task.taskStatus = Assigned
	task.assignedWorker = &workerUuid
	sr.tasksMap[workToAssign] = task

}
