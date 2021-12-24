package core

import (
	"colonies/pkg/crypto"
	"encoding/json"
	"strconv"
)

const (
	IN  int = 0
	OUT     = 1
	ERR     = 2
	ENV     = 4
)

type Attribute struct {
	ID            string `json:"attributeid"`
	TargetID      string `json:"targetid"`
	AttributeType int    `json:"attributetype"`
	Key           string `json:"key"`
	Value         string `json:"value"`
}

func CreateAttribute(targetID string, attributeType int, key string, value string) *Attribute {
	id := crypto.GenerateHash([]byte(targetID + key + strconv.Itoa(attributeType))).String()
	return &Attribute{ID: id,
		TargetID:      targetID,
		AttributeType: attributeType,
		Key:           key,
		Value:         value}
}

func ConvertJSONToAttribute(jsonString string) (*Attribute, error) {
	var attribute *Attribute
	err := json.Unmarshal([]byte(jsonString), &attribute)
	if err != nil {
		return nil, err
	}

	return attribute, nil
}

func (attribute *Attribute) SetValue(value string) {
	attribute.Value = value
}

func (attribute *Attribute) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(attribute, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
