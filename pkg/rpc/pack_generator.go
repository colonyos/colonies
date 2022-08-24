package rpc

import (
	"encoding/json"
)

const PackGeneratorPayloadType = "packgeneratormsg"

type PackGeneratorMsg struct {
	GeneratorID string `json:"generatorid"`
	Arg         string `json:"arg"`
	MsgType     string `json:"msgtype"`
}

func CreatePackGeneratorMsg(generatorID string, arg string) *PackGeneratorMsg {
	msg := &PackGeneratorMsg{}
	msg.GeneratorID = generatorID
	msg.Arg = arg
	msg.MsgType = PackGeneratorPayloadType

	return msg
}

func (msg *PackGeneratorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *PackGeneratorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *PackGeneratorMsg) Equals(msg2 *PackGeneratorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.GeneratorID == msg2.GeneratorID && msg.Arg == msg2.Arg {
		return true
	}

	return false
}

func CreatePackGeneratorMsgFromJSON(jsonString string) (*PackGeneratorMsg, error) {
	var msg *PackGeneratorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
