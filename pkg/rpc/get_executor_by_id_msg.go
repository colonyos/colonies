package rpc

import (
	"encoding/json"
)

const GetExecutorByIDPayloadType = "getexecutorbyidmsg"

type GetExecutorByIDMsg struct {
	ColonyName string `json:"colonyname"`
	ExecutorID string `json:"executorid"`
	MsgType    string `json:"msgtype"`
}

func CreateGetExecutorByIDMsg(colonyName string, executorID string) *GetExecutorByIDMsg {
	msg := &GetExecutorByIDMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorID = executorID
	msg.MsgType = GetExecutorByIDPayloadType

	return msg
}

func (msg *GetExecutorByIDMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetExecutorByIDMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetExecutorByIDMsg) Equals(msg2 *GetExecutorByIDMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ExecutorID == msg2.ExecutorID && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateGetExecutorByIDMsgFromJSON(jsonString string) (*GetExecutorByIDMsg, error) {
	var msg *GetExecutorByIDMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
