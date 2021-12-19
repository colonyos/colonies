package core

import "encoding/json"

const (
	PENDING  int = 0
	APPROVED     = 1
	REJECTED     = 2
)

type ComputerJSON struct {
	ID       string `json:"computerid"`
	Name     string `json:"name"`
	ColonyID string `json:"colonyid"`
	CPU      string `json:"cpu"`
	Cores    int    `json:"cores"`
	Mem      int    `json:"mem"`
	GPU      string `json:"gpu"`
	GPUs     int    `json:"gpus"`
	Status   int    `json:"status"`
}

type Computer struct {
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

func CreateComputer(id string, name string, colonyID string, cpu string, cores int, mem int, gpu string, gpus int) *Computer {
	return &Computer{id: id,
		name:     name,
		colonyID: colonyID,
		cpu:      cpu,
		cores:    cores,
		mem:      mem,
		gpu:      gpu,
		gpus:     gpus,
		status:   PENDING}
}

func CreateComputerFromDB(id string, name string, colonyID string, cpu string, cores int, mem int, gpu string, gpus int, status int) *Computer {
	return &Computer{id: id,
		name:     name,
		colonyID: colonyID,
		cpu:      cpu,
		cores:    cores,
		mem:      mem,
		gpu:      gpu,
		gpus:     gpus,
		status:   status}
}

func CreateComputerFromJSON(jsonString string) (*Computer, error) {
	var computerJSON ComputerJSON
	err := json.Unmarshal([]byte(jsonString), &computerJSON)
	if err != nil {
		return nil, err
	}

	return CreateComputerFromDB(computerJSON.ID, computerJSON.Name, computerJSON.ColonyID, computerJSON.CPU, computerJSON.Cores, computerJSON.Mem, computerJSON.GPU, computerJSON.GPUs, computerJSON.Status), nil
}

func CreateComputerArrayFromJSON(jsonString string) ([]*Computer, error) {
	var computers []*Computer
	var computersJSON []*ComputerJSON

	err := json.Unmarshal([]byte(jsonString), &computersJSON)
	if err != nil {
		return computers, err
	}

	for _, computerJSON := range computersJSON {
		computers = append(computers, CreateComputerFromDB(computerJSON.ID, computerJSON.Name, computerJSON.ColonyID, computerJSON.CPU, computerJSON.Cores, computerJSON.Mem, computerJSON.GPU, computerJSON.GPUs, computerJSON.Status))
	}

	return computers, nil
}

func ComputerArrayToJSON(computers []*Computer) (string, error) {
	var computersJSON []*ComputerJSON

	for _, computer := range computers {
		computerJSON := &ComputerJSON{ID: computer.id,
			Name:     computer.name,
			ColonyID: computer.colonyID,
			CPU:      computer.cpu,
			Cores:    computer.cores,
			Mem:      computer.mem,
			GPU:      computer.gpu,
			GPUs:     computer.gpus,
			Status:   computer.status}
		computersJSON = append(computersJSON, computerJSON)
	}

	jsonString, err := json.Marshal(computersJSON)
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

func (computer *Computer) ID() string {
	return computer.id
}

func (computer *Computer) Name() string {
	return computer.name
}

func (computer *Computer) ColonyID() string {
	return computer.colonyID
}

func (computer *Computer) CPU() string {
	return computer.cpu
}

func (computer *Computer) Cores() int {
	return computer.cores
}

func (computer *Computer) Mem() int {
	return computer.mem
}

func (computer *Computer) GPU() string {
	return computer.gpu
}

func (computer *Computer) GPUs() int {
	return computer.gpus
}

func (computer *Computer) Status() int {
	return computer.status
}

func (computer *Computer) IsApproved() bool {
	if computer.status == APPROVED {
		return true
	}

	return false
}

func (computer *Computer) IsRejected() bool {
	if computer.status == REJECTED {
		return true
	}

	return false
}

func (computer *Computer) IsPending() bool {
	if computer.status == PENDING {
		return true
	}

	return false
}

func (computer *Computer) Approve() {
	computer.status = APPROVED
}

func (computer *Computer) Reject() {
	computer.status = REJECTED
}

func (computer *Computer) ToJSON() (string, error) {
	computerJSON := &ComputerJSON{ID: computer.id,
		Name:     computer.name,
		ColonyID: computer.colonyID,
		CPU:      computer.cpu,
		Cores:    computer.cores,
		Mem:      computer.mem,
		GPU:      computer.gpu,
		GPUs:     computer.gpus,
		Status:   computer.status}

	jsonString, err := json.Marshal(computerJSON)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
