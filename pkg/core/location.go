package core

import (
	"encoding/json"
)

// Location represents a deployment site where executors can be bound
type Location struct {
	ID          string  `json:"locationid"`
	Name        string  `json:"name"`
	ColonyName  string  `json:"colonyname"`
	Description string  `json:"description"`
	Long        float64 `json:"long"`
	Lat         float64 `json:"lat"`
}

func CreateLocation(id string, name string, colonyName string, description string, long float64, lat float64) *Location {
	return &Location{
		ID:          id,
		Name:        name,
		ColonyName:  colonyName,
		Description: description,
		Long:        long,
		Lat:         lat,
	}
}

func ConvertJSONToLocation(jsonString string) (*Location, error) {
	var location *Location
	err := json.Unmarshal([]byte(jsonString), &location)
	if err != nil {
		return nil, err
	}

	return location, nil
}

func ConvertJSONToLocationArray(jsonString string) ([]*Location, error) {
	var locations []*Location

	err := json.Unmarshal([]byte(jsonString), &locations)
	if err != nil {
		return locations, err
	}

	return locations, nil
}

func ConvertLocationArrayToJSON(locations []*Location) (string, error) {
	jsonBytes, err := json.Marshal(locations)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func IsLocationArraysEqual(locations1 []*Location, locations2 []*Location) bool {
	counter := 0
	for _, loc1 := range locations1 {
		for _, loc2 := range locations2 {
			if loc1.Equals(loc2) {
				counter++
			}
		}
	}

	if counter == len(locations1) && counter == len(locations2) {
		return true
	}

	return false
}

func (location *Location) Equals(location2 *Location) bool {
	if location2 == nil {
		return false
	}

	if location.ID == location2.ID &&
		location.Name == location2.Name &&
		location.ColonyName == location2.ColonyName &&
		location.Description == location2.Description &&
		location.Long == location2.Long &&
		location.Lat == location2.Lat {
		return true
	}

	return false
}

func (location *Location) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(location)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
