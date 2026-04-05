package rpc

import (
	"encoding/json"
)

const GetAllMetricsPayloadType = "getallmetricsmsg"

type GetAllMetricsMsg struct {
	ColonyName   string `json:"colonyname"`
	ExecutorName string `json:"executorname"`
	MsgType      string `json:"msgtype"`
}

func CreateGetAllMetricsMsg(colonyName string, executorName string) *GetAllMetricsMsg {
	msg := &GetAllMetricsMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.MsgType = GetAllMetricsPayloadType
	return msg
}

func (msg *GetAllMetricsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *GetAllMetricsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *GetAllMetricsMsg) Equals(msg2 *GetAllMetricsMsg) bool {
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

func CreateGetAllMetricsMsgFromJSON(jsonString string) (*GetAllMetricsMsg, error) {
	var msg *GetAllMetricsMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}
