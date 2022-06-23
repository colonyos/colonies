package core

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/security/crypto"

	"github.com/google/uuid"
)

type RecurrentWorkflowSpec struct {
	ID           string `json:"generatorid"`
	Name         string `json:"name"`
	WorkflowSpec string `json:"workflowspec"`
	CronSpec     string `json:"cronspec"`
}

func CreateRecurrentWorkflowSpec(name string,
	workflowSpec string,
	cronSpec string) *RecurrentWorkflowSpec {
	uuid := uuid.New()
	crypto := crypto.CreateCrypto()
	id := crypto.GenerateHash(uuid.String())

	recWorkflowSpec := &RecurrentWorkflowSpec{
		ID:           id,
		Name:         name,
		WorkflowSpec: workflowSpec,
		CronSpec:     cronSpec,
	}

	return recWorkflowSpec
}

func ConvertJSONToRecurrentWorkflowSpec(jsonString string) (*RecurrentWorkflowSpec, error) {
	var recWorkflowSpec *RecurrentWorkflowSpec
	err := json.Unmarshal([]byte(jsonString), &recWorkflowSpec)
	if err != nil {
		return nil, err
	}

	return recWorkflowSpec, nil
}

func ConvertRecurrentWorkflowSpecArrayToJSON(recWorkflowSpecs []*RecurrentWorkflowSpec) (string, error) {
	jsonBytes, err := json.MarshalIndent(recWorkflowSpecs, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToRecurrentWorkflowSpecArray(jsonString string) ([]*RecurrentWorkflowSpec, error) {
	var recWorkflowSpecs []*RecurrentWorkflowSpec
	err := json.Unmarshal([]byte(jsonString), &recWorkflowSpecs)
	if err != nil {
		return recWorkflowSpecs, err
	}

	return recWorkflowSpecs, nil
}

func IsRecurrentWorkflowSpecArraysEqual(recWorkflowSpecs1 []*RecurrentWorkflowSpec,
	recWorkflowSpecs2 []*RecurrentWorkflowSpec) bool {
	if recWorkflowSpecs1 == nil || recWorkflowSpecs2 == nil {
		return false
	}

	counter := 0
	for _, recWorkflowSpec1 := range recWorkflowSpecs1 {
		for _, recWorkflowSpec2 := range recWorkflowSpecs2 {
			if recWorkflowSpec1.Equals(recWorkflowSpec2) {
				counter++
			}
		}
	}

	if counter == len(recWorkflowSpecs1) && counter == len(recWorkflowSpecs2) {
		return true
	}

	return false
}

func (recWorkflowSpec *RecurrentWorkflowSpec) Equals(recWorkflowSpec2 *RecurrentWorkflowSpec) bool {
	if recWorkflowSpec2 == nil {
		return false
	}

	same := true
	if recWorkflowSpec.ID != recWorkflowSpec2.ID ||
		recWorkflowSpec.Name != recWorkflowSpec2.Name ||
		recWorkflowSpec.WorkflowSpec != recWorkflowSpec2.WorkflowSpec ||
		recWorkflowSpec.CronSpec != recWorkflowSpec2.CronSpec {
		same = false
	}

	return same
}

func (recWorkflowSpec *RecurrentWorkflowSpec) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(recWorkflowSpec, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
