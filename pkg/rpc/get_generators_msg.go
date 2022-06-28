package rpc

import (
	"encoding/json"
)

const GetGeneratorsPayloadType = "getgeneratorsmsg"

type GetGeneratorsMsg struct {
	ColonyID string `json:"colonyid"`
	Count    int    `json:"count"`
	MsgType  string `json:"msgtype"`
}

func CreateGetGeneratorsMsg(colonyID string, count int) *GetGeneratorsMsg {
	msg := &GetGeneratorsMsg{}
	msg.ColonyID = colonyID
	msg.Count = count
	msg.MsgType = GetGeneratorsPayloadType

	return msg
}

func (msg *GetGeneratorsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetGeneratorsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetGeneratorsMsg) Equals(msg2 *GetGeneratorsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyID == msg2.ColonyID &&
		msg.Count == msg2.Count {
		return true
	}

	return false
}

func CreateGetGeneratorsMsgFromJSON(jsonString string) (*GetGeneratorsMsg, error) {
	var msg *GetGeneratorsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
