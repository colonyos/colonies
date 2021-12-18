package core

import "encoding/json"

const (
	PENDING  int = 0
	APPROVED     = 1
	REJECTED     = 2
)

type WorkerJSON struct {
	ID       string `json:"workerid"`
	Name     string `json:"name"`
	ColonyID string `json:"colonyid"`
	CPU      string `json:"cpu"`
	Cores    int    `json:"cores"`
	Mem      int    `json:"mem"`
	GPU      string `json:"gpu"`
	GPUs     int    `json:"gpus"`
	Status   int    `json:"status"`
}

type Worker struct {
	id       string
	name     string
	colonyID string
	cpu      string
	cores    int
	mem      int
	gpu      string
	gpus     int
	status   int
}

func CreateWorker(id string, name string, colonyID string, cpu string, cores int, mem int, gpu string, gpus int) *Worker {
	return &Worker{id: id, name: name, colonyID: colonyID, cpu: cpu, cores: cores, mem: mem, gpu: gpu, gpus: gpus, status: PENDING}
}

func CreateWorkerFromDB(id string, name string, colonyID string, cpu string, cores int, mem int, gpu string, gpus int, status int) *Worker {
	return &Worker{id: id, name: name, colonyID: colonyID, cpu: cpu, cores: cores, mem: mem, gpu: gpu, gpus: gpus, status: status}
}

func CreateWorkerFromJSON(jsonString string) (*Worker, error) {
	var workerJSON WorkerJSON
	err := json.Unmarshal([]byte(jsonString), &workerJSON)
	if err != nil {
		return nil, err
	}

	return CreateWorkerFromDB(workerJSON.ID, workerJSON.Name, workerJSON.ColonyID, workerJSON.CPU, workerJSON.Cores, workerJSON.Mem, workerJSON.GPU, workerJSON.GPUs, workerJSON.Status), nil
}

func CreateWorkerArrayFromJSON(jsonString string) ([]*Worker, error) {
	var workers []*Worker
	var workersJSON []*WorkerJSON

	err := json.Unmarshal([]byte(jsonString), &workersJSON)
	if err != nil {
		return workers, err
	}

	for _, workerJSON := range workersJSON {
		workers = append(workers, CreateWorkerFromDB(workerJSON.ID, workerJSON.Name, workerJSON.ColonyID, workerJSON.CPU, workerJSON.Cores, workerJSON.Mem, workerJSON.GPU, workerJSON.GPUs, workerJSON.Status))
	}

	return workers, nil
}

func WorkerArrayToJSON(workers []*Worker) (string, error) {
	var workersJSON []*WorkerJSON

	for _, worker := range workers {
		workerJSON := &WorkerJSON{ID: worker.ID(), Name: worker.Name(), ColonyID: worker.ColonyID(), CPU: worker.CPU(), Cores: worker.Cores(), Mem: worker.Mem(), GPU: worker.GPU(), GPUs: worker.GPUs(), Status: worker.Status()}
		workersJSON = append(workersJSON, workerJSON)
	}

	jsonString, err := json.Marshal(workersJSON)
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

func (worker *Worker) ID() string {
	return worker.id
}

func (worker *Worker) Name() string {
	return worker.name
}

func (worker *Worker) ColonyID() string {
	return worker.colonyID
}

func (worker *Worker) CPU() string {
	return worker.cpu
}

func (worker *Worker) Cores() int {
	return worker.cores
}

func (worker *Worker) Mem() int {
	return worker.mem
}

func (worker *Worker) GPU() string {
	return worker.gpu
}

func (worker *Worker) GPUs() int {
	return worker.gpus
}

func (worker *Worker) Status() int {
	return worker.status
}

func (worker *Worker) IsApproved() bool {
	if worker.status == APPROVED {
		return true
	}

	return false
}

func (worker *Worker) IsRejected() bool {
	if worker.status == REJECTED {
		return true
	}

	return false
}

func (worker *Worker) IsPending() bool {
	if worker.status == PENDING {
		return true
	}

	return false
}

func (worker *Worker) Approve() {
	worker.status = APPROVED
}

func (worker *Worker) Reject() {
	worker.status = REJECTED
}

func (worker *Worker) ToJSON() (string, error) {
	workerJSON := &WorkerJSON{ID: worker.ID(), Name: worker.Name(), ColonyID: worker.ColonyID(), CPU: worker.CPU(), Cores: worker.Cores(), Mem: worker.Mem(), GPU: worker.GPU(), GPUs: worker.GPUs(), Status: worker.Status()}

	jsonString, err := json.Marshal(workerJSON)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
