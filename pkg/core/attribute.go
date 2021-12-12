package core

import (
	"colonies/pkg/crypto"
	"strconv"
)

const (
	IN  int = 0
	OUT     = 1
	ERR     = 2
)

type Attribute struct {
	attributeID   string
	taskID        string
	attributeType int
	key           string
	value         string
}

func CreateAttribute(taskID string, attributeType int, key string, value string) *Attribute {
	attributeID := crypto.GenerateHash([]byte(taskID + key + strconv.Itoa(attributeType))).String()
	return &Attribute{attributeID: attributeID, taskID: taskID, attributeType: attributeType, key: key, value: value}
}

func (attribute *Attribute) ID() string {
	return attribute.attributeID
}

func (attribute *Attribute) TaskID() string {
	return attribute.taskID
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
