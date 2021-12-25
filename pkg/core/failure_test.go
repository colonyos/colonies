package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFailure(t *testing.T) {
	status := 200
	errorMessage := "error_msg"
	failure := CreateFailure(status, errorMessage)

	assert.Equal(t, status, failure.Status)
	assert.Equal(t, errorMessage, failure.Message)
}

func TestFailureEquals(t *testing.T) {
	failure1 := CreateFailure(200, "error_msg")
	failure2 := CreateFailure(201, "error_msg")
	failure3 := CreateFailure(200, "error_msg3")

	assert.True(t, failure1.Equals(failure1))
	assert.False(t, failure1.Equals(failure2))
	assert.False(t, failure1.Equals(failure3))
}

func TestFailureParseJSON(t *testing.T) {
	status := 200
	errorMessage := "error_msg"
	failure1 := CreateFailure(status, errorMessage)

	failureJSON, err := failure1.ToJSON()
	assert.Nil(t, err)

	failure2, err := ConvertJSONToFailure(failureJSON)
	assert.Nil(t, err)
	assert.True(t, failure2.Equals(failure1))
}
