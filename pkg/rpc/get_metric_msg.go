package rpc

import (
	"encoding/json"
)

const GetMetricPayloadType = "getmetricmsg"

type GetMetricMsg struct {
	ColonyName   string `json:"colonyname"`
	ExecutorName string `json:"executorname"`
	Key          string `json:"key"`
	MsgType      string `json:"msgtype"`
}

func CreateGetMetricMsg(colonyName string, executorName string, key string) *GetMetricMsg {
	msg := &GetMetricMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.Key = key
	msg.MsgType = GetMetricPayloadType
	return msg
}

func (msg *GetMetricMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *GetMetricMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *GetMetricMsg) Equals(msg2 *GetMetricMsg) bool {
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

func CreateGetMetricMsgFromJSON(jsonString string) (*GetMetricMsg, error) {
	var msg *GetMetricMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}
