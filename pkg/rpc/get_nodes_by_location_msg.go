package rpc

import (
	"encoding/json"
)

const GetNodesByLocationPayloadType = "getnodesbylocationmsg"

type GetNodesByLocationMsg struct {
	ColonyName string `json:"colonyname"`
	Location   string `json:"location"`
	MsgType    string `json:"msgtype"`
}

func CreateGetNodesByLocationMsg(colonyName string, location string) *GetNodesByLocationMsg {
	msg := &GetNodesByLocationMsg{}
	msg.ColonyName = colonyName
	msg.Location = location
	msg.MsgType = GetNodesByLocationPayloadType

	return msg
}

func (msg *GetNodesByLocationMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetNodesByLocationMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetNodesByLocationMsg) Equals(msg2 *GetNodesByLocationMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName && msg.Location == msg2.Location {
		return true
	}

	return false
}

func CreateGetNodesByLocationMsgFromJSON(jsonString string) (*GetNodesByLocationMsg, error) {
	var msg *GetNodesByLocationMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
