package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAttribute(t *testing.T) {
	key := "test_key"
	value := "test_value"

	attribute := CreateAttribute(GenerateRandomID(), OUT, key, value)

	assert.Len(t, attribute.ID, 64)
	assert.Equal(t, OUT, attribute.AttributeType)
	assert.Equal(t, key, attribute.Key)
	assert.Equal(t, value, attribute.Value)
}

func TestAttributeToJSON(t *testing.T) {
	key := "test_key"
	value := "test_value"

	attribute1 := CreateAttribute(GenerateRandomID(), OUT, key, value)
	jsonString, err := attribute1.ToJSON()
	assert.Nil(t, err)

	attribute2, err := ConvertJSONToAttribute(jsonString)
	assert.True(t, attribute2.Equals(attribute1))
}
