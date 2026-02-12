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

type Edge struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Animated bool   `json:"animated"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Data struct {
	Label string `json:"label"`
}

type Style struct {
	Color      string `json:"color"`
	Background string `json:"background"`
}

type GraphNode struct {
	ID       string   `json:"id"`
	Data     Data     `json:"data"`
	Position Position `json:"position"`
	Type     string   `json:"type"`
	Style    Style    `json:"style"`
}

type ProcessGraph struct {
	storage        ProcessGraphStorage
	ID             string    `json:"processgraphid"`
	InitiatorID    string    `json:"initiatorid"`
	InitiatorName  string    `json:"initiatorname"`
	ColonyName     string    `json:"colonyname"`
	Roots          []string  `json:"rootprocessids"`
	State          int       `json:"state"`
	SubmissionTime time.Time `json:"submissiontime"`
	StartTime      time.Time `json:"starttime"`
	EndTime        time.Time `json:"endtime"`
	ProcessIDs     []string  `json:"processids"`
	Nodes          []GraphNode    `json:"nodes"`
	Edges          []Edge    `json:"edges"`
	nodesMap       map[string]*GraphNode
	processCache   map[string]*Process
}

func CreateProcessGraph(colonyName string) (*ProcessGraph, error) {
	graph := &ProcessGraph{}
	graph.ColonyName = colonyName
	graph.Edges = make([]Edge, 0)
	graph.Nodes = make([]GraphNode, 0)

	uuid := uuid.New()
	crypto := crypto.CreateCrypto()
	id := crypto.GenerateHash(uuid.String())
	graph.ID = id

	graph.nodesMap = make(map[string]*GraphNode)

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
	jsonBytes, err := json.Marshal(processGraphs)
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
	graph.processCache = make(map[string]*Process)
}

// getProcess returns a process by ID, using a cache to avoid redundant DB queries.
// Within a single graph operation (e.g. Resolve, ToJSON), each process is fetched
// from the database at most once.
func (graph *ProcessGraph) getProcess(processID string) (*Process, error) {
	if graph.processCache != nil {
		if p, ok := graph.processCache[processID]; ok {
			return p, nil
		}
	}
	p, err := graph.storage.GetProcessByID(processID)
	if err != nil {
		return nil, err
	}
	if graph.processCache != nil && p != nil {
		graph.processCache[processID] = p
	}
	return p, nil
}

func (graph *ProcessGraph) Resolve() error {
	processes := 0
	failedProcesses := 0
	cancelledProcesses := 0
	runningProcesses := 0
	successfulProcesses := 0
	waitingProcesses := 0
	hasFailedParent := false
	hasCancelledParent := false

	// Pass 1: Count states, check parent dependencies, detect failures/cancellations
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
		if process.State == CANCELLED {
			cancelledProcesses++
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
			parent, err := graph.getProcess(parentProcessID)
			if err != nil {
				return err
			}
			if parent.State == SUCCESS {
				nrParentsFinished++
			} else if parent.State == FAILED {
				hasFailedParent = true
			} else if parent.State == CANCELLED {
				hasCancelledParent = true
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

	// Pass 2: If any parent failed, cascade failure to all processes in one pass
	if hasFailedParent {
		err = graph.Iterate(func(process *Process) error {
			if process.State != FAILED {
				process.State = FAILED
				if err := graph.storage.SetProcessState(process.ID, FAILED); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		graph.State = FAILED
		graph.storage.SetProcessGraphState(graph.ID, graph.State)
		return nil
	}

	// Pass 3: If any parent cancelled, cascade cancellation to all non-terminal processes
	if hasCancelledParent {
		err = graph.Iterate(func(process *Process) error {
			if process.State != CANCELLED && process.State != SUCCESS && process.State != FAILED {
				process.State = CANCELLED
				if err := graph.storage.SetProcessState(process.ID, CANCELLED); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		graph.State = CANCELLED
		graph.storage.SetProcessGraphState(graph.ID, graph.State)
		return nil
	}

	if failedProcesses >= 1 {
		graph.State = FAILED
	} else if cancelledProcesses >= 1 {
		graph.State = CANCELLED
	} else if successfulProcesses == processes {
		graph.State = SUCCESS
	} else if successfulProcesses > 0 || runningProcesses >= 1 {
		graph.State = RUNNING
	} else {
		graph.State = WAITING
	}

	graph.storage.SetProcessGraphState(graph.ID, graph.State)

	return nil
}

func (graph *ProcessGraph) GetRoot(childProcessID string) (*Process, error) {
	visited := make(map[string]bool)
	process, _, err := graph.getRoot(childProcessID, 0, visited)
	return process, err
}

func (graph *ProcessGraph) Depth(childProcessID string) (int, error) {
	visited := make(map[string]bool)
	_, counter, err := graph.getRoot(childProcessID, 0, visited)
	return counter, err
}

func (graph *ProcessGraph) calcEdges() error {
	if graph.storage == nil {
		return nil
	}
	err := graph.Iterate(func(process *Process) error {
		for _, child := range process.Children {
			id := process.ID + "-" + child
			source := process.ID
			target := child
			animated := false
			if process.State == RUNNING {
				animated = true
			}
			edge := Edge{ID: id, Source: source, Target: target, Animated: animated}
			graph.Edges = append(graph.Edges, edge)
		}

		return nil
	})

	return err
}

func (graph *ProcessGraph) calcNodes() error {
	if graph.storage == nil {
		return nil
	}

	paddingsPerLevel := make(map[int]int)
	nodesPerDepth := make(map[int][]*GraphNode)

	boxwidth := 150
	padding := 50

	err := graph.IterateWithDepth(func(process *Process, depth int) error {
		w, ok := paddingsPerLevel[depth]
		if ok {
			w = paddingsPerLevel[depth] + boxwidth + padding
			paddingsPerLevel[depth] = w
		} else {
			paddingsPerLevel[depth] = 0
		}

		x := w
		y := depth * 80
		t := ""
		if len(process.Parents) == 0 {
			t = "input"
		} else if len(process.Children) == 0 {
			t = "output"
		}

		background := "#eee8d8"
		switch process.State {
		case WAITING:
			background = "#eee8d8"
		case RUNNING:
			background = "#4689cd"
		case SUCCESS:
			background = "#92d050"
		case FAILED:
			background = "#cb4239"
		case CANCELLED:
			background = "#f5a623"
		}

		style := Style{Background: background}
		node := &GraphNode{ID: process.ID, Data: Data{Label: process.FunctionSpec.NodeName}, Position: Position{X: x, Y: y}, Type: t, Style: style}
		graph.nodesMap[process.ID] = node
		nodesPerDepth[depth] = append(nodesPerDepth[depth], node)
		return nil
	})

	maxWidth := 0
	for depth := range nodesPerDepth {
		paddingThisLevel, ok := paddingsPerLevel[depth]
		if ok {
			widthThisLevel := 0
			if paddingThisLevel == 0 {
				widthThisLevel = boxwidth
			} else {
				nrOfNodes := paddingThisLevel/(boxwidth+padding) + 1
				widthThisLevel = (nrOfNodes * boxwidth) + (padding * (nrOfNodes - 1))
			}
			if widthThisLevel > maxWidth {
				maxWidth = widthThisLevel
			}
		}
	}

	targetCenterPoint := maxWidth / 2

	for depth, nodes := range nodesPerDepth {
		paddingThisLevel, ok := paddingsPerLevel[depth]
		if ok {
			widthThisLevel := 0
			if paddingThisLevel == 0 {
				widthThisLevel = boxwidth
			} else {
				nrOfNodes := paddingThisLevel/(boxwidth+padding) + 1
				widthThisLevel = (nrOfNodes * boxwidth) + (padding * (nrOfNodes - 1))
			}

			if widthThisLevel < maxWidth {
				centerPoint := widthThisLevel / 2
				diff := targetCenterPoint - centerPoint
				for _, node := range nodes {
					node.Position.X = node.Position.X + diff
				}
			}

		}
	}

	for _, node := range graph.nodesMap {
		graph.Nodes = append(graph.Nodes, *node)
	}

	return err
}

func (graph *ProcessGraph) getRoot(childProcessID string, counter int, visited map[string]bool) (*Process, int, error) {
	process, err := graph.getProcess(childProcessID)
	if err != nil {
		return nil, -1, err
	}
	if visited[childProcessID] {
		return nil, -1, errors.New("Loops are not allowed in process graphs")
	}

	visited[childProcessID] = true

	if len(process.Parents) == 0 {
		return process, counter, nil
	} else {
		for _, childProcessID := range process.Parents {
			return graph.getRoot(childProcessID, counter+1, visited)
		}
	}

	return nil, counter, nil
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

func (graph *ProcessGraph) CancelledProcesses() (int, error) {
	counter := 0
	err := graph.Iterate(func(process *Process) error {
		if process.State == CANCELLED {
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
	graph.ProcessIDs = []string{}
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

func (graph *ProcessGraph) IterateWithDepth(visitFunc func(process *Process, depth int) error) error {
	visited := make(map[string]bool)
	for _, root := range graph.Roots {
		err := graph.iterateWithDepth(root, 0, visited, visitFunc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (graph *ProcessGraph) iterateWithDepth(processID string, depth int, visited map[string]bool, visitFunc func(process *Process, depth int) error) error {
	process, err := graph.getProcess(processID)
	if err != nil {
		return err
	}

	if process == nil {
		errMsg := "Failed to iterate processgraph, process with ID=" + processID + " not found"
		return errors.New(errMsg)
	}

	if visited[processID] {
		return nil
	}

	visited[processID] = true

	err = visitFunc(process, depth)
	if err != nil {
		return err
	}

	for _, childProcessID := range process.Children {
		err := graph.iterateWithDepth(childProcessID, depth+1, visited, visitFunc)
		if err != nil {
			return err
		}
	}

	return nil
}

func (graph *ProcessGraph) iterate(processID string, visited map[string]bool, visitFunc func(process *Process) error) error {
	process, err := graph.getProcess(processID)
	if err != nil {
		return err
	}

	if process == nil {
		errMsg := "Failed to iterate processgraph, process with ID=" + processID + " not found"
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
		graph.ColonyName == graph2.ColonyName &&
		graph.InitiatorID == graph2.InitiatorID &&
		graph.InitiatorName == graph2.InitiatorID {
		return true
	}

	return false
}

func (graph *ProcessGraph) ToJSON() (string, error) {
	err := graph.calcNodes()
	if err != nil {
		return "", err
	}

	err = graph.calcEdges()
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(graph)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
