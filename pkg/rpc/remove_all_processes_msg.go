package rpc

import (
	"encoding/json"
)

const RemoveAllProcessesPayloadType = "removeallprocessesmsg"

type RemoveAllProcessesMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
	State      int    `json:"state"`
}

func CreateRemoveAllProcessesMsg(colonyID string) *RemoveAllProcessesMsg {
	msg := &RemoveAllProcessesMsg{}
	msg.ColonyName = colonyID
	msg.MsgType = RemoveAllProcessesPayloadType

	return msg
}

func (msg *RemoveAllProcessesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveAllProcessesMsg) Equals(msg2 *RemoveAllProcessesMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName && msg.State == msg2.State {
		return true
	}

	return false
}

func (msg *RemoveAllProcessesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRemoveAllProcessesMsgFromJSON(jsonString string) (*RemoveAllProcessesMsg, error) {
	var msg *RemoveAllProcessesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
