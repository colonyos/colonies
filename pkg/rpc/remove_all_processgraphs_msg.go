package rpc

import (
	"encoding/json"
)

const RemoveAllProcessGraphsPayloadType = "removeallprocessgraphsmsg"

type RemoveAllProcessGraphsMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
	State      int    `json:"state"`
}

func CreateRemoveAllProcessGraphsMsg(colonyName string) *RemoveAllProcessGraphsMsg {
	msg := &RemoveAllProcessGraphsMsg{}
	msg.ColonyName = colonyName
	msg.MsgType = RemoveAllProcessGraphsPayloadType

	return msg
}

func (msg *RemoveAllProcessGraphsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveAllProcessGraphsMsg) Equals(msg2 *RemoveAllProcessGraphsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName && msg.State == msg2.State {
		return true
	}

	return false
}

func (msg *RemoveAllProcessGraphsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRemoveAllProcessGraphsMsgFromJSON(jsonString string) (*RemoveAllProcessGraphsMsg, error) {
	var msg *RemoveAllProcessGraphsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
