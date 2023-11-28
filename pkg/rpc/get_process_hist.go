package rpc

import (
	"encoding/json"
)

const GetProcessHistPayloadType = "getprocesshistmsg"

type GetProcessHistMsg struct {
	ColonyName string `json:"colonyname"`
	ExecutorID string `json:"executorid"`
	Seconds    int    `json:"seconds"`
	State      int    `json:"state"`
	MsgType    string `json:"msgtype"`
}

func CreateGetProcessHistMsg(colonyID string, executorID string, seconds int, state int) *GetProcessHistMsg {
	msg := &GetProcessHistMsg{}
	msg.ColonyName = colonyID
	msg.ExecutorID = executorID
	msg.Seconds = seconds
	msg.State = state
	msg.MsgType = GetProcessHistPayloadType

	return msg
}

func (msg *GetProcessHistMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessHistMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetProcessHistMsg) Equals(msg2 *GetProcessHistMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.ExecutorID == msg2.ExecutorID &&
		msg.Seconds == msg2.Seconds &&
		msg.State == msg2.State {
		return true
	}

	return false
}

func CreateGetProcessHistMsgFromJSON(jsonString string) (*GetProcessHistMsg, error) {
	var msg *GetProcessHistMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
