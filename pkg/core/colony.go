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

func IsColonyArraysEqual(colonies1 []*Colony, colonies2 []*Colony) bool {
	counter := 0
	for _, colony1 := range colonies1 {
		for _, colony2 := range colonies2 {
			if colony1.Equals(colony2) {
				counter++
			}
		}
	}

	if counter == len(colonies1) && counter == len(colonies2) {
		return true
	}

	return false
}

func (colony *Colony) SetID(id string) {
	colony.ID = id
}

func (colony *Colony) Equals(colony2 *Colony) bool {
	if colony.ID == colony2.ID &&
		colony.Name == colony2.Name {
		return true
	}

	return false
}

func (colony *Colony) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(colony)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
