package rpc

import (
	"encoding/json"
)

const IncGeneratorPayloadType = "incgeneratormsg"

type IncGeneratorMsg struct {
	GeneratorID string `json:"generatorid"`
	MsgType     string `json:"msgtype"`
}

func CreateIncGeneratorMsg(generatorID string) *IncGeneratorMsg {
	msg := &IncGeneratorMsg{}
	msg.GeneratorID = generatorID
	msg.MsgType = IncGeneratorPayloadType

	return msg
}

func (msg *IncGeneratorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *IncGeneratorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *IncGeneratorMsg) Equals(msg2 *IncGeneratorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.GeneratorID == msg2.GeneratorID {
		return true
	}

	return false
}

func CreateIncGeneratorMsgFromJSON(jsonString string) (*IncGeneratorMsg, error) {
	var msg *IncGeneratorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
