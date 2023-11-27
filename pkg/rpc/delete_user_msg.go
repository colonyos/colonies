package rpc

import (
	"encoding/json"
)

const DeleteUserPayloadType = "deleteusermsg"

type DeleteUserMsg struct {
	ColonyName string `json:"colonyname"`
	Name       string `json:"name"`
	MsgType    string `json:"msgtype"`
}

func CreateDeleteUserMsg(colonyName string, name string) *DeleteUserMsg {
	msg := &DeleteUserMsg{}
	msg.ColonyName = colonyName
	msg.Name = name
	msg.MsgType = DeleteUserPayloadType

	return msg
}

func (msg *DeleteUserMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteUserMsg) Equals(msg2 *DeleteUserMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.Name == msg2.Name && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func (msg *DeleteUserMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateDeleteUserMsgFromJSON(jsonString string) (*DeleteUserMsg, error) {
	var msg *DeleteUserMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
