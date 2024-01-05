package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsLogEquals(t *testing.T) {
	now := time.Now()
	log1 := &Log{ProcessID: "test_process_id", ColonyName: "test_colony", ExecutorName: "test_executor_name", Message: "test_msg", Timestamp: now.Unix()}
	log2 := &Log{ProcessID: "test_process_id", ColonyName: "test_colony", ExecutorName: "test_executor_name", Message: "test_msg", Timestamp: now.Unix()}

	assert.True(t, log1.Equals(log2))
	log1.Message = "changed_msg"
	assert.False(t, log1.Equals(log2))
}

func TestIsLogArraysEquals(t *testing.T) {
	now := time.Now()
	log1 := &Log{ProcessID: "test_process_id", ColonyName: "test_colony", ExecutorName: "test_executor_name", Message: "test_msg", Timestamp: now.Unix()}
	log2 := &Log{ProcessID: "test_process_id", ColonyName: "test_colony", ExecutorName: "test_executor_name", Message: "test_msg", Timestamp: now.Unix()}

	log3 := &Log{ProcessID: "test_process_id_2", ColonyName: "test_colony", ExecutorName: "test_executor_name", Message: "test_msg", Timestamp: now.Unix()}
	log4 := &Log{ProcessID: "test_process_id_2", ColonyName: "test_colony", ExecutorName: "test_executor_name", Message: "test_msg", Timestamp: now.Unix()}

	logs1 := []*Log{log1, log2}
	logs2 := []*Log{log3, log4}
	assert.True(t, IsLogArraysEqual(logs1, logs1))
	assert.False(t, IsLogArraysEqual(logs1, logs2))
}

func TestLogToJSON(t *testing.T) {
	now := time.Now()
	log1 := &Log{ProcessID: "test_process_id", ColonyName: "test_colony", ExecutorName: "test_executor_name", Message: "test_msg", Timestamp: now.Unix()}
	jsonStr, err := log1.ToJSON()
	assert.Nil(t, err)

	log2, err := ConvertJSONToLog(jsonStr)
	assert.Nil(t, err)
	assert.True(t, log1.Equals(log2))
}

func TestLogArrayToJSON(t *testing.T) {
	now := time.Now()
	log1 := &Log{ProcessID: "test_process_id", ColonyName: "test_colony", ExecutorName: "test_executor_name", Message: "test_msg", Timestamp: now.Unix()}
	log2 := &Log{ProcessID: "test_process_id", ColonyName: "test_colony", ExecutorName: "test_executor_name", Message: "test_msg", Timestamp: now.Unix()}

	logs1 := []*Log{log1, log2}

	jsonStr, err := ConvertLogArrayToJSON(logs1)
	assert.Nil(t, err)

	logs2, err := ConvertJSONToLogArray(jsonStr)
	assert.Nil(t, err)
	assert.True(t, IsLogArraysEqual(logs1, logs2))
}
