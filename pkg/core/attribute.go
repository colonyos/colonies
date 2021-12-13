package core

import (
	"colonies/pkg/crypto"
	"strconv"
)

const (
	IN  int = 0
	OUT     = 1
	ERR     = 2
	ENV     = 4
)

type Attribute struct {
	attributeID   string
	targetID      string
	attributeType int
	key           string
	value         string
}

func CreateAttribute(targetID string, attributeType int, key string, value string) *Attribute {
	attributeID := crypto.GenerateHash([]byte(targetID + key + strconv.Itoa(attributeType))).String()
	return &Attribute{attributeID: attributeID, targetID: targetID, attributeType: attributeType, key: key, value: value}
}

func (attribute *Attribute) ID() string {
	return attribute.attributeID
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

func (attribute *Attribute) ToMap(attributes []*Attribute) map[string]string {
	attributeMap := make(map[string]string)
	return attributeMap
}
