package rpc

import (
	"encoding/json"
)

const RemoveGeneratorPayloadType = "removegeneratormsg"

type RemoveGeneratorMsg struct {
	GeneratorID string `json:"generatorid"`
	MsgType     string `json:"msgtype"`
	All         bool   `json:"all"`
}

func CreateRemoveGeneratorMsg(generatorID string) *RemoveGeneratorMsg {
	msg := &RemoveGeneratorMsg{}
	msg.GeneratorID = generatorID
	msg.MsgType = RemoveGeneratorPayloadType

	return msg
}

func (msg *RemoveGeneratorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveGeneratorMsg) Equals(msg2 *RemoveGeneratorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.GeneratorID == msg2.GeneratorID && msg.All == msg2.All {
		return true
	}

	return false
}

func (msg *RemoveGeneratorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRemoveGeneratorMsgFromJSON(jsonString string) (*RemoveGeneratorMsg, error) {
	var msg *RemoveGeneratorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
