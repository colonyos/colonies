package core

import "encoding/json"

const (
	PENDING  int = 0
	APPROVED     = 1
	REJECTED     = 2
)

type Computer struct {
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

func CreateComputer(id string, name string, colonyID string, cpu string, cores int, mem int, gpu string, gpus int) *Computer {
	return &Computer{ID: id,
		Name:     name,
		ColonyID: colonyID,
		CPU:      cpu,
		Cores:    cores,
		Mem:      mem,
		GPU:      gpu,
		GPUs:     gpus,
		Status:   PENDING}
}

func CreateComputerFromDB(id string, name string, colonyID string, cpu string, cores int, mem int, gpu string, gpus int, status int) *Computer {
	computer := CreateComputer(id, name, colonyID, cpu, cores, mem, gpu, gpus)
	computer.Status = status
	return computer
}

func ConvertJSONToComputer(jsonString string) (*Computer, error) {
	var computer *Computer
	err := json.Unmarshal([]byte(jsonString), &computer)
	if err != nil {
		return nil, err
	}

	return computer, nil
}

func ConvertJSONToComputerArray(jsonString string) ([]*Computer, error) {
	var computers []*Computer
	err := json.Unmarshal([]byte(jsonString), &computers)
	if err != nil {
		return computers, err
	}

	return computers, nil
}

func ConvertComputerArrayToJSON(computers []*Computer) (string, error) {
	jsonBytes, err := json.MarshalIndent(computers, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (computer *Computer) IsApproved() bool {
	if computer.Status == APPROVED {
		return true
	}

	return false
}

func (computer *Computer) IsRejected() bool {
	if computer.Status == REJECTED {
		return true
	}

	return false
}

func (computer *Computer) IsPending() bool {
	if computer.Status == PENDING {
		return true
	}

	return false
}

func (computer *Computer) Approve() {
	computer.Status = APPROVED
}

func (computer *Computer) Reject() {
	computer.Status = REJECTED
}

func (computer *Computer) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(computer, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
