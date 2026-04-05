package rpc

import (
	"encoding/json"
)

const RemoveMetricPayloadType = "removemetricmsg"

type RemoveMetricMsg struct {
	ColonyName   string `json:"colonyname"`
	ExecutorName string `json:"executorname"`
	Key          string `json:"key"`
	MsgType      string `json:"msgtype"`
}

func CreateRemoveMetricMsg(colonyName string, executorName string, key string) *RemoveMetricMsg {
	msg := &RemoveMetricMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.Key = key
	msg.MsgType = RemoveMetricPayloadType
	return msg
}

func (msg *RemoveMetricMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *RemoveMetricMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *RemoveMetricMsg) Equals(msg2 *RemoveMetricMsg) bool {
	if msg2 == nil {
		return false
	}
	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.ExecutorName == msg2.ExecutorName &&
		msg.Key == msg2.Key {
		return true
	}
	return false
}

func CreateRemoveMetricMsgFromJSON(jsonString string) (*RemoveMetricMsg, error) {
	var msg *RemoveMetricMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}
