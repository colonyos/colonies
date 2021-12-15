package core

import "encoding/json"

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

func CreateColonyFromJSON(jsonString string) (*Colony, error) {
	var colonyJSON ColonyJSON
	err := json.Unmarshal([]byte(jsonString), &colonyJSON)
	if err != nil {
		return nil, err
	}

	return CreateColony(colonyJSON.ID, colonyJSON.Name), nil
}

func (colony *Colony) Name() string {
	return colony.name
}

func (colony *Colony) ID() string {
	return colony.id
}

func (colony *Colony) ToJSON() (string, error) {
	colonyJSON := &ColonyJSON{ID: colony.ID(), Name: colony.Name()}

	jsonString, err := json.Marshal(colonyJSON)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
