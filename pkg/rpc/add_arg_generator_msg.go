package rpc

import (
	"encoding/json"
)

const AddArgGeneratorPayloadType = "addarggeneratormsg"

type AddArgGeneratorMsg struct {
	GeneratorID string `json:"generatorid"`
	Arg         string `json:"arg"`
	MsgType     string `json:"msgtype"`
}

func CreateAddArgGeneratorMsg(generatorID string, arg string) *AddArgGeneratorMsg {
	msg := &AddArgGeneratorMsg{}
	msg.GeneratorID = generatorID
	msg.Arg = arg
	msg.MsgType = AddArgGeneratorPayloadType

	return msg
}

func (msg *AddArgGeneratorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddArgGeneratorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *AddArgGeneratorMsg) Equals(msg2 *AddArgGeneratorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.GeneratorID == msg2.GeneratorID && msg.Arg == msg2.Arg {
		return true
	}

	return false
}

func CreateAddArgGeneratorMsgFromJSON(jsonString string) (*AddArgGeneratorMsg, error) {
	var msg *AddArgGeneratorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
