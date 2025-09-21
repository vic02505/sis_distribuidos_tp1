package utils

func (sr *SharedResources) getFirstAvailableMappingTask() (*string, *Task) {

	var taskName string 
	var taskToDo Task 

	for fileSplit, task := range sr.tasksMap {
		if task.TaskStatus == NotAssigned && task.TaskType == Map {
			taskName = fileSplit
			taskToDo = task
			break
		}
	}

	return &taskName, &taskToDo
}

func (sr *SharedResources) getFirstAvailableReduceTask() (*string, *Task) {

	var taskName string 
	var taskToDo Task 

	for fileSplit, task := range sr.tasksMap {
		if task.TaskStatus == NotAssigned && task.TaskType == Reduce {
			taskName = fileSplit
			taskToDo = task
			break
		}
	}

	return &taskName, &taskToDo
}

func (sr *SharedResources) assignTask(workToAssign, workerUuid string) {
	task := sr.tasksMap[workToAssign]
	task.TaskStatus = Assigned
	task.AssignedWorker = &workerUuid
	sr.tasksMap[workToAssign] = task

}
