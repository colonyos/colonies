package rpc

import (
	"encoding/json"
)

const GetMetricsPayloadType = "getmetricsmsg"

type GetMetricsMsg struct {
	ColonyName   string `json:"colonyname"`
	ExecutorName string `json:"executorname"`
	MsgType      string `json:"msgtype"`
}

func CreateGetMetricsMsg(colonyName string, executorName string) *GetMetricsMsg {
	msg := &GetMetricsMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.MsgType = GetMetricsPayloadType
	return msg
}

func (msg *GetMetricsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *GetMetricsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *GetMetricsMsg) Equals(msg2 *GetMetricsMsg) bool {
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

func CreateGetMetricsMsgFromJSON(jsonString string) (*GetMetricsMsg, error) {
	var msg *GetMetricsMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}
