package core

import (
	"encoding/json"
)

type ColonyJSON struct {
	ID   string `json:"colonyid"`
	Name string `json:"name"`
}

type Colony struct {
	id   string
	name string
}

func CreateColony(id string, name string) *Colony {
	colony := &Colony{id: id, name: name}

	return colony
}

func ConvertJSONToColony(jsonString string) (*Colony, error) {
	var colonyJSON ColonyJSON
	err := json.Unmarshal([]byte(jsonString), &colonyJSON)
	if err != nil {
		return nil, err
	}

	return CreateColony(colonyJSON.ID, colonyJSON.Name), nil
}

func ConvertJSONToColonyArray(jsonString string) ([]*Colony, error) {
	var colonies []*Colony
	var coloniesJSON []*ColonyJSON

	err := json.Unmarshal([]byte(jsonString), &coloniesJSON)
	if err != nil {
		return colonies, err
	}

	for _, colonyJSON := range coloniesJSON {
		colonies = append(colonies, CreateColony(colonyJSON.ID, colonyJSON.Name))
	}

	return colonies, nil
}

func ConvertColonyArrayToJSON(colonies []*Colony) (string, error) {
	var coloniesJSON []ColonyJSON

	for _, colony := range colonies {
		colonyJSON := ColonyJSON{ID: colony.id, Name: colony.name}
		coloniesJSON = append(coloniesJSON, colonyJSON)
	}

	jsonString, err := json.MarshalIndent(coloniesJSON, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}

func (colony *Colony) Name() string {
	return colony.name
}

func (colony *Colony) SetID(id string) {
	colony.id = id
}

func (colony *Colony) ID() string {
	return colony.id
}

func (colony *Colony) ToJSON() (string, error) {
	colonyJSON := &ColonyJSON{ID: colony.id, Name: colony.name}

	jsonString, err := json.Marshal(colonyJSON)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
