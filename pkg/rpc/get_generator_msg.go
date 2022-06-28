package rpc

import (
	"encoding/json"
)

const GetGeneratorPayloadType = "getgeneratormsg"

type GetGeneratorMsg struct {
	GeneratorID string `json:"generatorid"`
	MsgType     string `json:"msgtype"`
}

func CreateGetGeneratorMsg(generatorID string) *GetGeneratorMsg {
	msg := &GetGeneratorMsg{}
	msg.GeneratorID = generatorID
	msg.MsgType = GetGeneratorPayloadType

	return msg
}

func (msg *GetGeneratorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetGeneratorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetGeneratorMsg) Equals(msg2 *GetGeneratorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.GeneratorID == msg2.GeneratorID {
		return true
	}

	return false
}

func CreateGetGeneratorMsgFromJSON(jsonString string) (*GetGeneratorMsg, error) {
	var msg *GetGeneratorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
