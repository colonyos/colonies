package rpc

import (
	"encoding/json"
)

const GetExecutorsPayloadType = "getexecutorsmsg"

type GetExecutorsMsg struct {
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
}

func CreateGetExecutorsMsg(colonyID string) *GetExecutorsMsg {
	msg := &GetExecutorsMsg{}
	msg.ColonyID = colonyID
	msg.MsgType = GetExecutorsPayloadType

	return msg
}

func (msg *GetExecutorsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetExecutorsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetExecutorsMsg) Equals(msg2 *GetExecutorsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID {
		return true
	}

	return false
}

func CreateGetExecutorsMsgFromJSON(jsonString string) (*GetExecutorsMsg, error) {
	var msg *GetExecutorsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
