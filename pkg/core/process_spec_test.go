package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessSpecJSON(t *testing.T) {
	colonyID := GenerateRandomID()
	computerType := "test_computer_type"
	timeout := -1
	maxRetries := 3
	mem := 1000
	cores := 10
	gpus := 1
	in := make(map[string]string)
	in["test_key"] = "test_value"
	processSpec := CreateProcessSpec(colonyID, []string{}, computerType, timeout, maxRetries, mem, cores, gpus, in)

	jsonString, err := processSpec.ToJSON()
	assert.Nil(t, err)

	fmt.Println(jsonString)

}
