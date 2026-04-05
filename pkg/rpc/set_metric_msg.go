package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const SetMetricPayloadType = "setmetricmsg"

type SetMetricMsg struct {
	Metric core.Metric `json:"metric"`
	MsgType string     `json:"msgtype"`
}

func CreateSetMetricMsg(metric core.Metric) *SetMetricMsg {
	msg := &SetMetricMsg{}
	msg.Metric = metric
	msg.MsgType = SetMetricPayloadType
	return msg
}

func (msg *SetMetricMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *SetMetricMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *SetMetricMsg) Equals(msg2 *SetMetricMsg) bool {
	if msg2 == nil {
		return false
	}
	if msg.MsgType == msg2.MsgType && msg.Metric.Equals(msg2.Metric) {
		return true
	}
	return false
}

func CreateSetMetricMsgFromJSON(jsonString string) (*SetMetricMsg, error) {
	var msg *SetMetricMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}
