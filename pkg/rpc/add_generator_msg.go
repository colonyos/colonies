package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const AddGeneratorPayloadType = "addgeneratormsg"

type AddGeneratorMsg struct {
	Generator *core.Generator `json:"generator"`
	MsgType   string          `json:"msgtype"`
}

func CreateAddGeneratorMsg(generator *core.Generator) *AddGeneratorMsg {
	msg := &AddGeneratorMsg{}
	msg.Generator = generator
	msg.MsgType = AddGeneratorPayloadType

	return msg
}

func (msg *AddGeneratorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddGeneratorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddGeneratorMsg) Equals(msg2 *AddGeneratorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.Generator.Equals(msg2.Generator) {
		return true
	}

	return false
}

func CreateAddGeneratorMsgFromJSON(jsonString string) (*AddGeneratorMsg, error) {
	var msg *AddGeneratorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
