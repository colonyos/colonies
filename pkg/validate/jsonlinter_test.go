package validate

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateJSON(t *testing.T) {
	jsonStr := `{
"rggee2": 12
"rggee": "ewwgewge"
}`

	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	assert.NotNil(t, err)
	if err != nil {
		jsonErrStr, _ := JSON(err, jsonStr, false)
		assert.True(t, len(jsonErrStr) > 0)
	}
}

func TestJSONNilError(t *testing.T) {
	result, err := JSON(nil, `{"test": 1}`, false)
	assert.Nil(t, err)
	assert.Equal(t, "", result)
}

func TestJSONFullMode(t *testing.T) {
	jsonStr := `{
"rggee2": 12
"rggee": "ewwgewge"
}`

	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	assert.NotNil(t, err)

	jsonErrStr, err2 := JSON(err, jsonStr, true)
	assert.Nil(t, err2)
	assert.Contains(t, jsonErrStr, prefixStr)
}

func TestJSONNonSyntaxError(t *testing.T) {
	customErr := errors.New("custom error")
	result, err := JSON(customErr, `{"test": 1}`, false)
	assert.Equal(t, "", result)
	assert.Equal(t, customErr, err)
}

func TestAddStringToLineSuccess(t *testing.T) {
	original := "line1\nline2\nline3"
	result, err := AddStringToLine(original, 2, "message")
	assert.Nil(t, err)
	assert.Contains(t, result, "line2"+prefixStr+"message")
}

func TestAddStringToLineOutOfRange(t *testing.T) {
	original := "line1\nline2"

	// Line number too high
	_, err := AddStringToLine(original, 5, "message")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "out of range")

	// Line number too low
	_, err = AddStringToLine(original, 0, "message")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "out of range")
}

func TestFindErrorLineNumber(t *testing.T) {
	// Non-syntax error should return -1
	customErr := errors.New("not a syntax error")
	lineNum := findErrorLineNumber(`{"test": 1}`, customErr)
	assert.Equal(t, -1, lineNum)

	// Syntax error should return line number
	jsonStr := `{
"key": 123
"missing_comma": true
}`
	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	assert.NotNil(t, err)
	lineNum = findErrorLineNumber(jsonStr, err)
	assert.Greater(t, lineNum, 0)
}
