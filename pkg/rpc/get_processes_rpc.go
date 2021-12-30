package rpc

import (
	"encoding/json"
)

const GetProcessesMsgType = "GetProcesses"

type GetProcessesMsg struct {
	RPC      RPC    `json:"rpc"`
	ColonyID string `json:"coloyid"`
	Count    int    `json:"count"`
	State    int    `json:"state"`
}

func CreateGetProcessesMsg(colonyID string, count int, state int) *GetProcessesMsg {
	msg := &GetProcessesMsg{}
	msg.RPC.Method = GetProcessesMsgType
	msg.ColonyID = colonyID
	msg.Count = count
	msg.State = state

	return msg
}

func (msg *GetProcessesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetProcessesMsgFromJSON(jsonString string) (*GetProcessesMsg, error) {
	var msg *GetProcessesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}