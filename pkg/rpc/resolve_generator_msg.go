package rpc

import (
	"encoding/json"
)

const ResolveGeneratorPayloadType = "resolvegeneratormsg"

type ResolveGeneratorMsg struct {
	ColonyName    string `json:"colonyname"`
	GeneratorName string `json:"generatorname"`
	MsgType       string `json:"msgtype"`
}

func CreateResolveGeneratorMsg(colonyName string, generatorName string) *ResolveGeneratorMsg {
	msg := &ResolveGeneratorMsg{}
	msg.GeneratorName = generatorName
	msg.ColonyName = colonyName
	msg.MsgType = ResolveGeneratorPayloadType

	return msg
}

func (msg *ResolveGeneratorMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ResolveGeneratorMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ResolveGeneratorMsg) Equals(msg2 *ResolveGeneratorMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.GeneratorName == msg2.GeneratorName && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateResolveGeneratorMsgFromJSON(jsonString string) (*ResolveGeneratorMsg, error) {
	var msg *ResolveGeneratorMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
