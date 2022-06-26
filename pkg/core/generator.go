package core

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/security/crypto"

	"github.com/google/uuid"
)

type Generator struct {
	ID           string `json:"generatorid"`
	ColonyID     string `json:"colonyid"`
	Name         string `json:"name"`
	WorkflowSpec string `json:"workflowspec"`
	Trigger      int    `json:"trigger"`
	Counter      int    `json:"counter"`
	Timeout      int    `json:"timeout"`
}

func CreateGenerator(colonyID string, name string, workflowSpec string, trigger int, counter int, timeout int) *Generator {
	uuid := uuid.New()
	crypto := crypto.CreateCrypto()
	id := crypto.GenerateHash(uuid.String())

	generator := &Generator{
		ID:           id,
		ColonyID:     colonyID,
		Name:         name,
		WorkflowSpec: workflowSpec,
		Trigger:      trigger,
		Counter:      counter,
		Timeout:      timeout,
	}

	return generator
}

func ConvertJSONToGenerator(jsonString string) (*Generator, error) {
	var generator *Generator
	err := json.Unmarshal([]byte(jsonString), &generator)
	if err != nil {
		return nil, err
	}

	return generator, nil
}

func ConvertGeneratorArrayToJSON(generators []*Generator) (string, error) {
	jsonBytes, err := json.MarshalIndent(generators, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToGeneratorArray(jsonString string) ([]*Generator, error) {
	var generators []*Generator
	err := json.Unmarshal([]byte(jsonString), &generators)
	if err != nil {
		return generators, err
	}

	return generators, nil
}

func IsGeneratorArraysEqual(generators1 []*Generator, generators2 []*Generator) bool {
	if generators1 == nil || generators2 == nil {
		return false
	}

	counter := 0
	for _, generator1 := range generators1 {
		for _, generator2 := range generators2 {
			if generator1.Equals(generator2) {
				counter++
			}
		}
	}

	if counter == len(generators1) && counter == len(generators2) {
		return true
	}

	return false
}

func (generator *Generator) Equals(generator2 *Generator) bool {
	if generator2 == nil {
		return false
	}

	same := true
	if generator.ID != generator2.ID ||
		generator.ColonyID != generator2.ColonyID ||
		generator.Name != generator2.Name ||
		generator.WorkflowSpec != generator2.WorkflowSpec ||
		generator.Trigger != generator2.Trigger ||
		generator.Counter != generator2.Counter ||
		generator.Timeout != generator2.Timeout {
		same = false
	}

	return same
}

func (generator *Generator) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(generator, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
