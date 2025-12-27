package rpc

import (
	"encoding/json"
)

const GetLocationPayloadType = "getlocationmsg"

type GetLocationMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
	Name       string `json:"name"`
}

func CreateGetLocationMsg(colonyName string, name string) *GetLocationMsg {
	msg := &GetLocationMsg{}
	msg.MsgType = GetLocationPayloadType
	msg.ColonyName = colonyName
	msg.Name = name

	return msg
}

func (msg *GetLocationMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetLocationMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetLocationMsg) Equals(msg2 *GetLocationMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName && msg.Name == msg2.Name {
		return true
	}

	return false
}

func CreateGetLocationMsgFromJSON(jsonString string) (*GetLocationMsg, error) {
	var msg *GetLocationMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
