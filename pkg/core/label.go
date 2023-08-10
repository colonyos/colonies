package core

import (
	"encoding/json"
)

type Label struct {
	Name  string `json:"name"`
	Files int    `json:"files"`
}

func ConvertJSONToLabel(jsonString string) (*Label, error) {
	var label *Label
	err := json.Unmarshal([]byte(jsonString), &label)
	if err != nil {
		return &Label{}, err
	}

	return label, nil
}

func ConvertLabelArrayToJSON(labels []*Label) (string, error) {
	jsonBytes, err := json.Marshal(labels)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToLabelArray(jsonString string) ([]*Label, error) {
	var labels []*Label
	err := json.Unmarshal([]byte(jsonString), &labels)
	if err != nil {
		return labels, err
	}

	return labels, nil
}

func (label *Label) Equals(label2 *Label) bool {
	same := true
	if label.Name != label2.Name ||
		label.Files != label2.Files {
		same = false
	}

	return same
}

func IsLabelArraysEqual(labels1 []*Label, labels2 []*Label) bool {
	if labels1 == nil || labels2 == nil {
		return false
	}

	if len(labels1) != len(labels2) {
		return false
	}

	counter := 0
	for i := range labels1 {
		if labels1[i].Equals(labels2[i]) {
			counter++
		}
	}

	if counter == len(labels1) {
		return true
	}

	return false
}

func (label *Label) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(label)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
