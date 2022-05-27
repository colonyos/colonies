package core

import "errors"

type ProcessGraphStorage interface {
	GetProcessByID(processID string) (*Process, error)
}

type ProcessGraph struct {
	storage ProcessGraphStorage
}

func CreateProcessGraph(storage ProcessGraphStorage) *ProcessGraph {
	graph := &ProcessGraph{}
	graph.storage = storage

	return graph
}

func (graph *ProcessGraph) EnableChildren(process *Process) {
}

func (graph *ProcessGraph) GetRoot(processID string) (*Process, error) {
	visited := make(map[string]bool)
	return graph.getRoot(processID, visited)
}

func (graph *ProcessGraph) getRoot(processID string, visited map[string]bool) (*Process, error) {
	process, err := graph.storage.GetProcessByID(processID)
	if err != nil {
		return nil, err
	}
	if visited[processID] {
		return nil, errors.New("loops are not allowed in process graphs")
	}

	visited[processID] = true

	if len(process.Parents) == 0 {
		return process, nil
	} else {
		for _, childProcessID := range process.Parents {
			return graph.getRoot(childProcessID, visited)
		}
	}

	return nil, nil
}

func (graph *ProcessGraph) CloseGraphAsFailed(process *Process) {
}
