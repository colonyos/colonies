package core

import (
	"encoding/json"
)

type Colony struct {
	ID   string `json:"colonyid"`
	Name string `json:"name"`
}

func CreateColony(id string, name string) *Colony {
	colony := &Colony{ID: id, Name: name}

	return colony
}

func ConvertJSONToColony(jsonString string) (*Colony, error) {
	var colony *Colony
	err := json.Unmarshal([]byte(jsonString), &colony)
	if err != nil {
		return nil, err
	}

	return colony, nil
}

func ConvertJSONToColonyArray(jsonString string) ([]*Colony, error) {
	var colonies []*Colony

	err := json.Unmarshal([]byte(jsonString), &colonies)
	if err != nil {
		return colonies, err
	}

	return colonies, nil
}

func ConvertColonyArrayToJSON(colonies []*Colony) (string, error) {
	jsonBytes, err := json.MarshalIndent(colonies, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (colony *Colony) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(colony)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
