package utils

func (sr *SharedResources) getFirstAvailableMappingTask() string {

	var availableTask string

	for fileSplit, task := range sr.tasksMap {
		if task.taskStatus == NotAssigned {
			availableTask = fileSplit
			break
		}
	}

	return availableTask
}

func (sr *SharedResources) isThereAvailableMappingTasks() bool {
	for _, task := range sr.tasksMap {
		if task.taskStatus == NotAssigned {
			return true
		}
	}
	return false
}

/*
func (sr *SharedResources) isThereAvailableReducingTasks() bool {
	for _, state := range sr.ReduceTasks {
		if state == NotAssigned {
			return true
		}
	}
	return false
}
*/

func (sr *SharedResources) isThereAvailableWork() bool {
	return sr.isThereAvailableMappingTasks()
}

func (sr *SharedResources) assignMappingWork(workToAssign, workerUuid string) {
	task := sr.tasksMap[workToAssign]
	task.taskStatus = Assigned
	task.assignedWorker = &workerUuid
	sr.tasksMap[workToAssign] = task

}
