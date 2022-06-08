package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateAttribute(t *testing.T) {
	key := "test_key"
	value := "test_value"

	attribute := CreateAttribute(GenerateRandomID(), GenerateRandomID(), GenerateRandomID(), OUT, key, value)

	assert.Len(t, attribute.ID, 64)
	assert.Equal(t, OUT, attribute.AttributeType)
	assert.Equal(t, key, attribute.Key)
	assert.Equal(t, value, attribute.Value)
}

func TestIsAttributeEquals(t *testing.T) {
	attribute1 := CreateAttribute(GenerateRandomID(), GenerateRandomID(), GenerateRandomID(), OUT, "test_key", "test_value")
	attribute2 := CreateAttribute(GenerateRandomID(), GenerateRandomID(), "", OUT, "test_key", "test_value")
	attribute3 := CreateAttribute(attribute1.ID, GenerateRandomID(), "", IN, "test_key", "test_value")
	attribute4 := CreateAttribute(attribute1.ID, GenerateRandomID(), GenerateRandomID(), ERR, "test_key", "test_value")
	attribute5 := CreateAttribute(attribute1.ID, GenerateRandomID(), GenerateRandomID(), OUT, "test_keyX", "test_value")
	attribute6 := CreateAttribute(attribute1.ID, GenerateRandomID(), "", OUT, "test_key", "test_valueX")

	assert.True(t, attribute1.Equals(attribute1))
	assert.False(t, attribute1.Equals(attribute2))
	assert.False(t, attribute1.Equals(attribute3))
	assert.False(t, attribute1.Equals(attribute4))
	assert.False(t, attribute1.Equals(attribute5))
	assert.False(t, attribute1.Equals(attribute6))
	assert.False(t, attribute1.Equals(nil))
}

func TestIsAttributeArraysEqual(t *testing.T) {
	attribute1 := CreateAttribute(GenerateRandomID(), GenerateRandomID(), "", OUT, "test_key", "test_value")
	attribute2 := CreateAttribute(GenerateRandomID(), GenerateRandomID(), GenerateRandomID(), OUT, "test_key", "test_value")
	attribute3 := CreateAttribute(GenerateRandomID(), GenerateRandomID(), "", OUT, "test_key", "test_value")
	attribute4 := CreateAttribute(GenerateRandomID(), GenerateRandomID(), GenerateRandomID(), OUT, "test_key", "test_value")

	var attributes1 []*Attribute
	attributes1 = append(attributes1, attribute1)
	attributes1 = append(attributes1, attribute2)
	attributes1 = append(attributes1, attribute3)

	var attributes2 []*Attribute
	attributes2 = append(attributes2, attribute2)
	attributes2 = append(attributes2, attribute3)
	attributes2 = append(attributes2, attribute1)

	var attributes3 []*Attribute
	attributes3 = append(attributes3, attribute2)
	attributes3 = append(attributes3, attribute3)
	attributes3 = append(attributes3, attribute4)

	var attributes4 []*Attribute

	assert.True(t, IsAttributeArraysEqual(attributes1, attributes1))
	assert.True(t, IsAttributeArraysEqual(attributes1, attributes2))
	assert.False(t, IsAttributeArraysEqual(attributes1, attributes3))
	assert.False(t, IsAttributeArraysEqual(attributes1, attributes4))
	assert.True(t, IsAttributeArraysEqual(attributes4, attributes4))
	assert.True(t, IsAttributeArraysEqual(nil, nil))
	assert.False(t, IsAttributeArraysEqual(nil, attributes2))
}

func TestAttributeToJSON(t *testing.T) {
	key := "test_key"
	value := "test_value"

	attribute1 := CreateAttribute(GenerateRandomID(), GenerateRandomID(), "", OUT, key, value)
	jsonString, err := attribute1.ToJSON()
	assert.Nil(t, err)

	attribute2, err := ConvertJSONToAttribute(jsonString + "error")
	assert.NotNil(t, err)

	attribute2, err = ConvertJSONToAttribute(jsonString)
	assert.Nil(t, err)
	assert.True(t, attribute2.Equals(attribute1))
}
