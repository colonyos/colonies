package rpc

import (
	"encoding/json"
)

const GetProcessHistPayloadType = "getprocesshistmsg"

type GetProcessHistMsg struct {
	ColonyID  string `json:"colonyid"`
	RuntimeID string `json:"runtimeid"`
	Seconds   int    `json:"seconds"`
	State     int    `json:"state"`
	MsgType   string `json:"msgtype"`
}

func CreateGetProcessHistMsg(colonyID string, runtimeID string, seconds int, state int) *GetProcessHistMsg {
	msg := &GetProcessHistMsg{}
	msg.ColonyID = colonyID
	msg.RuntimeID = runtimeID
	msg.Seconds = seconds
	msg.State = state
	msg.MsgType = GetProcessHistPayloadType

	return msg
}

func (msg *GetProcessHistMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessHistMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetProcessHistMsgFromJSON(jsonString string) (*GetProcessHistMsg, error) {
	var msg *GetProcessHistMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
