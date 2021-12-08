package core

const (
	PENDING  int = 0
	APPROVED     = 1
	REJECTED     = 2
)

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
