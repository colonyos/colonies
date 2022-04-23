package rpc

import (
	"encoding/json"
)

const GetProcessStatPayloadType = "getprocstatmsg"

type GetProcessStatMsg struct {
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
}

func CreateGetProcessStatMsg(colonyID string) *GetProcessStatMsg {
	msg := &GetProcessStatMsg{}
	msg.ColonyID = colonyID
	msg.MsgType = GetProcessStatPayloadType

	return msg
}

func (msg *GetProcessStatMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessStatMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessStatMsg) Equals(msg2 *GetProcessStatMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID {
		return true
	}

	return false
}

func CreateGetProcessStatMsgFromJSON(jsonString string) (*GetProcessStatMsg, error) {
	var msg *GetProcessStatMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
