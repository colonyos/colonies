package rpc

import (
	"encoding/json"
)

const GetProcessGraphsPayloadType = "getprocessgraphsmsg"

type GetProcessGraphsMsg struct {
	ColonyName string `json:"colonyname"`
	Count      int    `json:"count"`
	State      int    `json:"state"`
	MsgType    string `json:"msgtype"`
}

func CreateGetProcessGraphsMsg(colonyName string, count int, state int) *GetProcessGraphsMsg {
	msg := &GetProcessGraphsMsg{}
	msg.ColonyName = colonyName
	msg.Count = count
	msg.State = state
	msg.MsgType = GetProcessGraphsPayloadType

	return msg
}

func (msg *GetProcessGraphsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessGraphsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessGraphsMsg) Equals(msg2 *GetProcessGraphsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.Count == msg2.Count &&
		msg.State == msg2.State {
		return true
	}

	return false
}

func CreateGetProcessGraphsMsgFromJSON(jsonString string) (*GetProcessGraphsMsg, error) {
	var msg *GetProcessGraphsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
