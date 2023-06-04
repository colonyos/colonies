package rpc

import (
	"encoding/json"
)

const GetProcessesPayloadType = "getprocessesmsg"

type GetProcessesMsg struct {
	ColonyID     string `json:"colonyid"`
	Count        int    `json:"count"`
	State        int    `json:"state"`
	ExecutorType string `json:"executortype"`
	MsgType      string `json:"msgtype"`
}

func CreateGetProcessesMsg(colonyID string, count int, state int, executorType string) *GetProcessesMsg {
	msg := &GetProcessesMsg{}
	msg.ColonyID = colonyID
	msg.Count = count
	msg.State = state
	msg.ExecutorType = executorType
	msg.MsgType = GetProcessesPayloadType

	return msg
}

func (msg *GetProcessesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessesMsg) Equals(msg2 *GetProcessesMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyID == msg2.ColonyID &&
		msg.Count == msg2.Count &&
		msg.State == msg2.State &&
		msg.ExecutorType == msg2.ExecutorType {
		return true
	}

	return false
}

func CreateGetProcessesMsgFromJSON(jsonString string) (*GetProcessesMsg, error) {
	var msg *GetProcessesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
