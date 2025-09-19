package utils

func (sr *SharedResources) getFirstAvailableMappingTask() string {

	var availableTask string

	for fileSplit, state := range sr.MappingTasks {
		if state == NotAssigned {
			availableTask = fileSplit
			break
		}
	}

	return availableTask
}

func (sr *SharedResources) isThereAvailableMappingTasks() bool {
	for _, state := range sr.MappingTasks {
		if state == NotAssigned {
			return true
		}
	}
	return false
}

func (sr *SharedResources) isThereAvailableReducingTasks() bool {
	for _, state := range sr.ReduceTasks {
		if state == NotAssigned {
			return true
		}
	}
	return false
}
