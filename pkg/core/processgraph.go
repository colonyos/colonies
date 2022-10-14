package core

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type ProcessGraphStorage interface {
	GetProcessByID(processID string) (*Process, error)
	SetProcessState(processID string, state int) error
	SetWaitForParents(processID string, waitForParent bool) error
	SetProcessGraphState(processGraphID string, state int) error
}

type ProcessGraph struct {
	storage        ProcessGraphStorage
	ID             string    `json:"processgraphid"`
	ColonyID       string    `json:"colonyid"`
	Roots          []string  `json:"rootprocessids"`
	State          int       `json:"state"`
	SubmissionTime time.Time `json:"submissiontime"`
	StartTime      time.Time `json:"starttime"`
	EndTime        time.Time `json:"endtime"`
	ProcessIDs     []string  `json:"processids"`
}

func CreateProcessGraph(colonyID string) (*ProcessGraph, error) {
	graph := &ProcessGraph{}
	graph.ColonyID = colonyID

	uuid := uuid.New()
	crypto := crypto.CreateCrypto()
	id := crypto.GenerateHash(uuid.String())

	graph.ID = id

	return graph, nil
}

func ConvertJSONToProcessGraphWithStorage(jsonString string) (*ProcessGraph, error) {
	var processGraph *ProcessGraph
	err := json.Unmarshal([]byte(jsonString), &processGraph)
	if err != nil {
		return nil, err
	}

	return processGraph, nil
}

func ConvertJSONToProcessGraph(jsonString string) (*ProcessGraph, error) {
	var processGraph *ProcessGraph
	err := json.Unmarshal([]byte(jsonString), &processGraph)
	if err != nil {
		return nil, err
	}

	return processGraph, nil
}

func ConvertProcessGraphArrayToJSON(processGraphs []*ProcessGraph) (string, error) {
	jsonBytes, err := json.MarshalIndent(processGraphs, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToProcessGraphArray(jsonString string) ([]*ProcessGraph, error) {
	var processGraphs []*ProcessGraph
	err := json.Unmarshal([]byte(jsonString), &processGraphs)
	if err != nil {
		return processGraphs, err
	}

	return processGraphs, nil
}

func IsProcessGraphArraysEqual(processGraphs1 []*ProcessGraph, processGraphs2 []*ProcessGraph) bool {
	counter := 0
	for _, processGraph1 := range processGraphs1 {
		for _, processGraph2 := range processGraphs2 {
			if processGraph1.Equals(processGraph2) {
				counter++
			}
		}
	}

	if counter == len(processGraphs1) && counter == len(processGraphs2) {
		return true
	}

	return false
}

func (graph *ProcessGraph) AddRoot(processID string) {
	graph.Roots = append(graph.Roots, processID)
}

func (graph *ProcessGraph) SetStorage(storage ProcessGraphStorage) {
	graph.storage = storage
}

// Note: This function requires a working graph.storage reference
func (graph *ProcessGraph) Resolve() error {
	processes := 0
	failedProcesses := 0
	runningProcesses := 0
	successfulProcesses := 0
	waitingProcesses := 0

	err := graph.Iterate(func(process *Process) error {
		if process == nil {
			errMsg := "Failed to iterate processgraph, process is nil"
			log.Error(errMsg)
			return errors.New(errMsg)
		}
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
			process.WaitForParents = false
			graph.storage.SetWaitForParents(process.ID, false)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if failedProcesses >= 1 {
		graph.State = FAILED
	} else if successfulProcesses == processes {
		graph.State = SUCCESS
	} else if runningProcesses >= 1 {
		graph.State = RUNNING
	} else {
		graph.State = WAITING
	}

	graph.storage.SetProcessGraphState(graph.ID, graph.State)

	return err
}

func (graph *ProcessGraph) GetRoot(childProcessID string) (*Process, error) {
	visited := make(map[string]bool)
	return graph.getRoot(childProcessID, visited)
}

func (graph *ProcessGraph) getRoot(childProcessID string, visited map[string]bool) (*Process, error) {
	process, err := graph.storage.GetProcessByID(childProcessID)
	if err != nil {
		return nil, err
	}
	if visited[childProcessID] {
		return nil, errors.New("Loops are not allowed in process graphs")
	}

	visited[childProcessID] = true

	if len(process.Parents) == 0 {
		return process, nil
	} else {
		for _, childProcessID := range process.Parents {
			return graph.getRoot(childProcessID, visited)
		}
	}

	return nil, nil
}

func (graph *ProcessGraph) Leaves() ([]string, error) {
	var leafs []string

	err := graph.Iterate(func(process *Process) error {
		if len(process.Children) == 0 {
			leafs = append(leafs, process.ID)
		}
		return nil
	})
	return leafs, err
}

func (graph *ProcessGraph) Processes() (int, error) {
	counter := 0
	err := graph.Iterate(func(process *Process) error {
		counter++
		return nil
	})
	return counter, err
}

func (graph *ProcessGraph) WaitProcesses() (int, error) {
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

func (graph *ProcessGraph) WaitForParents() (int, error) {
	counter := 0
	err := graph.Iterate(func(process *Process) error {
		if process.WaitForParents {
			counter++
		}
		return nil
	})
	return counter, err
}

func (graph *ProcessGraph) UpdateProcessIDs() error {
	return graph.Iterate(func(process *Process) error {
		graph.ProcessIDs = append(graph.ProcessIDs, process.ID)
		return nil
	})
}

func (graph *ProcessGraph) Iterate(visitFunc func(process *Process) error) error {
	visited := make(map[string]bool)
	var err error

	for _, root := range graph.Roots {
		err = graph.iterate(root, visited, visitFunc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (graph *ProcessGraph) iterate(processID string, visited map[string]bool, visitFunc func(process *Process) error) error {
	process, err := graph.storage.GetProcessByID(processID)
	if err != nil {
		return err
	}

	if process == nil {
		errMsg := "Failed to iterate processgraph, process is nil"
		log.Error(errMsg)
		return errors.New(errMsg)
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
		graph.ColonyID == graph2.ColonyID {
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
