package rpc

import (
	"encoding/json"
)

const RemoveAllMetricsPayloadType = "removeallmetricsmsg"

type RemoveAllMetricsMsg struct {
	ColonyName   string `json:"colonyname"`
	ExecutorName string `json:"executorname"`
	MsgType      string `json:"msgtype"`
}

func CreateRemoveAllMetricsMsg(colonyName string, executorName string) *RemoveAllMetricsMsg {
	msg := &RemoveAllMetricsMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.MsgType = RemoveAllMetricsPayloadType
	return msg
}

func (msg *RemoveAllMetricsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *RemoveAllMetricsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *RemoveAllMetricsMsg) Equals(msg2 *RemoveAllMetricsMsg) bool {
	if msg2 == nil {
		return false
	}
	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.ExecutorName == msg2.ExecutorName {
		return true
	}
	return false
}

func CreateRemoveAllMetricsMsgFromJSON(jsonString string) (*RemoveAllMetricsMsg, error) {
	var msg *RemoveAllMetricsMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}
