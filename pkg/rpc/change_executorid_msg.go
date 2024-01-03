package rpc

import (
	"encoding/json"
)

const ChangeExecutorIDPayloadType = "changeexecutoridmsg"

type ChangeExecutorIDMsg struct {
	ColonyName string `json:"colonyname"`
	ExecutorID string `json:"executorid"`
	MsgType    string `json:"msgtype"`
}

func CreateChangeExecutorIDMsg(colonyName string, executorID string) *ChangeExecutorIDMsg {
	msg := &ChangeExecutorIDMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorID = executorID
	msg.MsgType = ChangeExecutorIDPayloadType

	return msg
}

func (msg *ChangeExecutorIDMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ChangeExecutorIDMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ChangeExecutorIDMsg) Equals(msg2 *ChangeExecutorIDMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ExecutorID == msg2.ExecutorID && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateChangeExecutorIDMsgFromJSON(jsonString string) (*ChangeExecutorIDMsg, error) {
	var msg *ChangeExecutorIDMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
