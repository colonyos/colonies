package rpc

import (
	"encoding/json"
)

const GetNodePayloadType = "getnodemsg"

type GetNodeMsg struct {
	ColonyName string `json:"colonyname"`
	NodeName   string `json:"nodename"`
	MsgType    string `json:"msgtype"`
}

func CreateGetNodeMsg(colonyName string, nodeName string) *GetNodeMsg {
	msg := &GetNodeMsg{}
	msg.ColonyName = colonyName
	msg.NodeName = nodeName
	msg.MsgType = GetNodePayloadType

	return msg
}

func (msg *GetNodeMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetNodeMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetNodeMsg) Equals(msg2 *GetNodeMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.NodeName == msg2.NodeName && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateGetNodeMsgFromJSON(jsonString string) (*GetNodeMsg, error) {
	var msg *GetNodeMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
