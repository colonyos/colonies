package rpc

import (
	"encoding/json"
)

const GetLogsPayloadType = "getlogsmsg"

type GetLogsMsg struct {
	ColonyName   string `json:"colonyname"`
	ProcessID    string `json:"processid"`
	ExecutorName string `json:"executorname"`
	Count        int    `json:"count"`
	Since        int64  `json:"since"`
	Latest       bool   `json:"latest"` // If true, return latest logs (descending order)
	MsgType      string `json:"msgtype"`
}

func CreateGetLogsMsg(colonyName string, processID string, count int, since int64) *GetLogsMsg {
	msg := &GetLogsMsg{}
	msg.ColonyName = colonyName
	msg.ProcessID = processID
	msg.Count = count
	msg.Since = since
	msg.MsgType = GetLogsPayloadType

	return msg
}

func (msg *GetLogsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetLogsMsg) Equals(msg2 *GetLogsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.ProcessID == msg2.ProcessID &&
		msg.Count == msg2.Count &&
		msg.Since == msg2.Since &&
		msg.ExecutorName == msg2.ExecutorName &&
		msg.Latest == msg2.Latest {
		return true
	}

	return false
}

func (msg *GetLogsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetLogsMsgFromJSON(jsonString string) (*GetLogsMsg, error) {
	var msg *GetLogsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
