package validate

import (
	"encoding/json"
	"fmt"
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
		fmt.Println(jsonErrStr)
		assert.True(t, len(jsonErrStr) > 0)
	}

}
