package rpc

import (
	"encoding/json"
)

const GetNodesPayloadType = "getnodesmsg"

type GetNodesMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
}

func CreateGetNodesMsg(colonyName string) *GetNodesMsg {
	msg := &GetNodesMsg{}
	msg.ColonyName = colonyName
	msg.MsgType = GetNodesPayloadType

	return msg
}

func (msg *GetNodesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetNodesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetNodesMsg) Equals(msg2 *GetNodesMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateGetNodesMsgFromJSON(jsonString string) (*GetNodesMsg, error) {
	var msg *GetNodesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
