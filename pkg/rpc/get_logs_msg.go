package rpc

import (
	"encoding/json"
)

const GetLogsPayloadType = "getlogsmsg"

type GetLogsMsg struct {
	ProcessID string `json:"processid"`
	Count     int    `json:"count"`
	MsgType   string `json:"msgtype"`
}

func CreateGetLogsMsg(processID string, count int) *GetLogsMsg {
	msg := &GetLogsMsg{}
	msg.ProcessID = processID
	msg.Count = count
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

	if msg.MsgType == msg2.MsgType && msg.ProcessID == msg2.ProcessID && msg.Count == msg2.Count {
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
