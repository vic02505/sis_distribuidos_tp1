package utils

import (
	"sync"
)

type SharedResources struct {
	mutex            sync.Mutex
	MappingTasks     map[string]string
	AssignedMapper   map[string]string
	ReduceTasks      map[string]string
	AssignedReducers map[string]string
}

func CreateInitialSharedResources(fileSplits []string) *SharedResources {

	mapperTasks := make(map[string]string)

	for _, fileSplit := range fileSplits {
		mapperTasks[fileSplit] = NotAssigned
	}

	return &SharedResources{
		MappingTasks:     mapperTasks,
		AssignedMapper:   make(map[string]string),
		ReduceTasks:      make(map[string]string),
		AssignedReducers: make(map[string]string),
	}
}

func (sr *SharedResources) AssignMappingWork(workerId string) string {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	availableTask := sr.getFirstAvailableMappingTask()

	sr.AssignedMapper[workerId] = availableTask
	return availableTask
}

func (sr *SharedResources) IsThereAvailableWork() bool {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()

	if len(sr.MappingTasks) > 0 && len(sr.ReduceTasks) == 0 {
		return sr.isThereAvailableMappingTasks()
	} else if len(sr.ReduceTasks) > 0 && len(sr.MappingTasks) == 0 {
		return sr.isThereAvailableReducingTasks()
	}

	return false
}

func (sr *SharedResources) MarkWorkAsFinished(workerUuid string, workType string) {
	sr.mutex.Lock()
	defer sr.mutex.Unlock()
	sr.AssignedReducers[workerUuid] = Finished
}
