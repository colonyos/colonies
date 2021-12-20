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

type AttributeJSON struct {
	ID            string `json:"attributeid"`
	TargetID      string `json:"targetid"`
	AttributeType int    `json:"attributetype"`
	Key           string `json:"key"`
	Value         string `json:"value"`
}

type Attribute struct {
	id            string
	targetID      string
	attributeType int
	key           string
	value         string
}

func CreateAttribute(targetID string, attributeType int, key string, value string) *Attribute {
	id := crypto.GenerateHash([]byte(targetID + key + strconv.Itoa(attributeType))).String()
	return &Attribute{id: id,
		targetID:      targetID,
		attributeType: attributeType,
		key:           key,
		value:         value}
}

func ConvertJSONToAttribute(jsonString string) (*Attribute, error) {
	var attributeJSON AttributeJSON
	err := json.Unmarshal([]byte(jsonString), &attributeJSON)
	if err != nil {
		return nil, err
	}

	return CreateAttribute(attributeJSON.TargetID, attributeJSON.AttributeType, attributeJSON.Key, attributeJSON.Value), nil
}

func convertToAttributeJSON(attributes []*Attribute) []*AttributeJSON {
	var attributesJSON []*AttributeJSON
	for _, attribute := range attributes {
		attributesJSON = append(attributesJSON, &AttributeJSON{ID: attribute.id,
			TargetID:      attribute.targetID,
			AttributeType: attribute.attributeType,
			Key:           attribute.key,
			Value:         attribute.value})
	}

	return attributesJSON
}

func convertFromAttributeJSON(attributesJSON []*AttributeJSON) []*Attribute {
	var attributes []*Attribute
	for _, attributeJSON := range attributesJSON {
		attributes = append(attributes, &Attribute{id: attributeJSON.ID,
			targetID:      attributeJSON.TargetID,
			attributeType: attributeJSON.AttributeType,
			key:           attributeJSON.Key,
			value:         attributeJSON.Value})
	}

	return attributes
}

func (attribute *Attribute) ID() string {
	return attribute.id
}

func (attribute *Attribute) TargetID() string {
	return attribute.targetID
}

func (attribute *Attribute) AttributeType() int {
	return attribute.attributeType
}

func (attribute *Attribute) Key() string {
	return attribute.key
}

func (attribute *Attribute) Value() string {
	return attribute.value
}

func (attribute *Attribute) SetValue(value string) {
	attribute.value = value
}

func (attribute *Attribute) ToJSON() (string, error) {
	attributeJSON := &AttributeJSON{ID: attribute.id,
		TargetID:      attribute.targetID,
		AttributeType: attribute.attributeType,
		Key:           attribute.key,
		Value:         attribute.value}

	jsonString, err := json.Marshal(attributeJSON)
	if err != nil {
		return "", err
	}

	return string(jsonString), nil
}
