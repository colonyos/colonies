package core

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/google/uuid"
)

type ProcessGraphStorage interface {
	GetProcessByID(processID string) (*Process, error)
	SetProcessState(processID string, state int) error
	SetWaitingForParents(processID string, waitingForParent bool) error
	SetProcessGraphState(processGraphID string, state int) error
}

type ProcessGraph struct {
	storage        ProcessGraphStorage
	ID             string    `json:"processgraphid"`
	Root           string    `json:"rootprocessid"`
	State          int       `json:"state"`
	SubmissionTime time.Time `json:"submissiontime"`
	EndTime        time.Time `json:"endtime"`
	RuntimeGroup   string    `json:"runtimegroup"`
}

func CreateProcessGraph(storage ProcessGraphStorage, rootProcessID string) (*ProcessGraph, error) {
	graph := &ProcessGraph{}
	graph.storage = storage
	graph.Root = rootProcessID

	uuid := uuid.New()
	crypto := crypto.CreateCrypto()
	id := crypto.GenerateHash(uuid.String())

	graph.ID = id

	return graph, nil
}

func ConvertJSONToProcessGraph(jsonString string, storage ProcessGraphStorage) (*ProcessGraph, error) {
	var processGraph *ProcessGraph
	err := json.Unmarshal([]byte(jsonString), &processGraph)
	if err != nil {
		return nil, err
	}

	processGraph.storage = storage
	return processGraph, nil
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
				// Set all processes in the graph as failed if one process fails
				err := graph.Iterate(func(process *Process) error {
					process.State = FAILED
					err = graph.storage.SetProcessState(process.ID, FAILED)
					if err != nil {
						return err
					}
					return nil
				})
				if err != nil {
					return err
				}
			}
		}
		if nrParentsFinished == nrParents {
			process.WaitingForParents = false
			graph.storage.SetWaitingForParents(process.ID, false)
		}
		return nil
	})

	if failedProcesses > 1 {
		graph.State = FAILED
	} else if successfulProcesses == processes {
		graph.State = SUCCESS
	} else if runningProcesses > 1 {
		graph.State = RUNNING
	} else {
		graph.State = WAITING
	}

	graph.storage.SetProcessGraphState(graph.ID, graph.State)

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

func (graph *ProcessGraph) Equals(graph2 *ProcessGraph) bool {
	if graph.State == graph2.State &&
		graph.ID == graph2.ID &&
		graph.EndTime.Unix() == graph2.EndTime.Unix() &&
		graph.RuntimeGroup == graph2.RuntimeGroup {
		return true
	}

	return false
}

func (graph *ProcessGraph) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(graph, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
