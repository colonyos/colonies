package rpc

import (
	"encoding/json"
)

const GetRuntimesPayloadType = "getruntimesmsg"

type GetRuntimesMsg struct {
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
}

func CreateGetRuntimesMsg(colonyID string) *GetRuntimesMsg {
	msg := &GetRuntimesMsg{}
	msg.ColonyID = colonyID
	msg.MsgType = GetRuntimesPayloadType

	return msg
}

func (msg *GetRuntimesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetRuntimesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetRuntimesMsg) Equals(msg2 *GetRuntimesMsg) bool {
	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID {
		return true
	}

	return false
}

func CreateGetRuntimesMsgFromJSON(jsonString string) (*GetRuntimesMsg, error) {
	var msg *GetRuntimesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
