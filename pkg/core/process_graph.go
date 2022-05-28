package core

import (
	"errors"
)

type ProcessGraphStorage interface {
	GetProcessByID(processID string) (*Process, error)
}

type ProcessGraph struct {
	storage ProcessGraphStorage
	Root    string
	State   int
}

func CreateProcessGraph(storage ProcessGraphStorage, rootProcessID string) (*ProcessGraph, error) {
	graph := &ProcessGraph{}
	graph.storage = storage
	graph.Root = rootProcessID

	return graph, nil
}

func (graph *ProcessGraph) Resolve() error {
	processes := 0
	failedProcesses := 0
	runningProcesses := 0
	successfulProcesses := 0
	waitingProcesses := 0

	err := graph.Iterate(func(process *Process) error {
		nrParents := len(process.Parents)
		nrParentsFinished := 0

		processes++
		if process.State == FAILED {
			failedProcesses++
		}
		if process.State == RUNNING {
			runningProcesses++
		}
		if process.State == WAITING {
			waitingProcesses++
		}
		if process.State == SUCCESS {
			successfulProcesses++
		}

		for _, parentProcessID := range process.Parents {
			parent, err := graph.storage.GetProcessByID(parentProcessID)
			if err != nil {
				return err
			}
			if parent.State == SUCCESS {
				nrParentsFinished++
			} else if parent.State == FAILED {
				// Set all process in graph as failed
				err := graph.Iterate(func(process *Process) error {
					process.State = FAILED
					// TODO: update database
					return nil
				})
				if err != nil {
					return err
				}
			}
		}
		if nrParentsFinished == nrParents {
			process.WaitingForParents = false
			// TODO: update database
		}
		return nil
	})

	if failedProcesses > 1 {
		graph.State = FAILED
	} else if successfulProcesses == processes {
		graph.State = SUCCESS
	} else if runningProcesses > 1 {
		graph.State = RUNNING
	} else if waitingProcesses > 1 {
		graph.State = WAITING
	}

	// TODO: update database

	return err
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

func (graph *ProcessGraph) Processes() (int, error) {
	counter := 0
	err := graph.Iterate(func(process *Process) error {
		counter++
		return nil
	})
	return counter, err
}

func (graph *ProcessGraph) WaitingProcesses() (int, error) {
	counter := 0
	err := graph.Iterate(func(process *Process) error {
		if process.State == WAITING {
			counter++
		}
		return nil
	})
	return counter, err
}

func (graph *ProcessGraph) RunningProcesses() (int, error) {
	counter := 0
	err := graph.Iterate(func(process *Process) error {
		if process.State == RUNNING {
			counter++
		}
		return nil
	})
	return counter, err
}

func (graph *ProcessGraph) SuccessfulProcesses() (int, error) {
	counter := 0
	err := graph.Iterate(func(process *Process) error {
		if process.State == SUCCESS {
			counter++
		}
		return nil
	})
	return counter, err
}

func (graph *ProcessGraph) FailedProcesses() (int, error) {
	counter := 0
	err := graph.Iterate(func(process *Process) error {
		if process.State == FAILED {
			counter++
		}
		return nil
	})
	return counter, err
}

func (graph *ProcessGraph) WaitingForParents() (int, error) {
	counter := 0
	err := graph.Iterate(func(process *Process) error {
		if process.WaitingForParents {
			counter++
		}
		return nil
	})
	return counter, err
}

func (graph *ProcessGraph) Iterate(visitFunc func(process *Process) error) error {
	visited := make(map[string]bool)
	return graph.iterate(graph.Root, visited, visitFunc)
}

func (graph *ProcessGraph) iterate(processID string, visited map[string]bool, visitFunc func(process *Process) error) error {
	process, err := graph.storage.GetProcessByID(processID)
	if err != nil {
		return err
	}
	if visited[processID] {
		return nil
	}

	visited[processID] = true

	err = visitFunc(process)
	if err != nil {
		return err
	}

	if len(process.Children) == 0 {
		return nil
	} else {
		for _, childProcessID := range process.Children {
			err := graph.iterate(childProcessID, visited, visitFunc)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
