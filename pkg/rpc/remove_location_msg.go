package rpc

import (
	"encoding/json"
)

const RemoveLocationPayloadType = "removelocationmsg"

type RemoveLocationMsg struct {
	ColonyName string `json:"colonyname"`
	Name       string `json:"name"`
	MsgType    string `json:"msgtype"`
}

func CreateRemoveLocationMsg(colonyName string, name string) *RemoveLocationMsg {
	msg := &RemoveLocationMsg{}
	msg.ColonyName = colonyName
	msg.Name = name
	msg.MsgType = RemoveLocationPayloadType

	return msg
}

func (msg *RemoveLocationMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveLocationMsg) Equals(msg2 *RemoveLocationMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.Name == msg2.Name && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func (msg *RemoveLocationMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRemoveLocationMsgFromJSON(jsonString string) (*RemoveLocationMsg, error) {
	var msg *RemoveLocationMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
