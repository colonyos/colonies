package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAttribute(t *testing.T) {
	key := "test_key"
	value := "test_value"

	attribute := CreateAttribute(GenerateRandomID(), OUT, key, value)

	assert.Len(t, attribute.ID(), 64)
	assert.Equal(t, OUT, attribute.AttributeType())
	assert.Equal(t, key, attribute.Key())
	assert.Equal(t, value, attribute.Value())
}

func TestAttributeToJSON(t *testing.T) {
	key := "test_key"
	value := "test_value"

	attribute := CreateAttribute(GenerateRandomID(), OUT, key, value)
	jsonString, err := attribute.ToJSON()
	assert.Nil(t, err)

	attribute2, err := CreateAttributeFromJSON(jsonString)

	assert.Equal(t, attribute.ID(), attribute2.ID())
	assert.Equal(t, attribute.TargetID(), attribute2.TargetID())
	assert.Equal(t, attribute.AttributeType(), attribute2.AttributeType())
	assert.Equal(t, attribute.Key(), attribute2.Key())
	assert.Equal(t, attribute.Value(), attribute2.Value())
}
