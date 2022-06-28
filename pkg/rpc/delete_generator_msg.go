package rpc

import (
	"encoding/json"
)

const DeleteGeneratorPayloadType = "deletegeneratormsg"

type DeleteGeneratorMsg struct {
	GeneratorID string `json:"generatorid"`
	MsgType     string `json:"msgtype"`
	All         bool   `json:"all"`
}

func CreateDeleteGeneratorMsg(generatorID string) *DeleteGeneratorMsg {
	msg := &DeleteGeneratorMsg{}
	msg.GeneratorID = generatorID
	msg.MsgType = DeleteGeneratorPayloadType

	return msg
}

func (msg *DeleteGeneratorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteGeneratorMsg) Equals(msg2 *DeleteGeneratorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.GeneratorID == msg2.GeneratorID && msg.All == msg2.All {
		return true
	}

	return false
}

func (msg *DeleteGeneratorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateDeleteGeneratorMsgFromJSON(jsonString string) (*DeleteGeneratorMsg, error) {
	var msg *DeleteGeneratorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
