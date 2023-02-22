package rpc

import (
	"encoding/json"
)

const GetFunctionsPayloadType = "getfunctionsmsg"

type GetFunctionsMsg struct {
	ExecutorID string `json:"executorid"`
	MsgType    string `json:"msgtype"`
}

func CreateGetFunctionsMsg(executorID string) *GetFunctionsMsg {
	msg := &GetFunctionsMsg{}
	msg.ExecutorID = executorID
	msg.MsgType = GetFunctionsPayloadType

	return msg
}

func (msg *GetFunctionsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetFunctionsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetFunctionsMsg) Equals(msg2 *GetFunctionsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ExecutorID == msg2.ExecutorID {
		return true
	}

	return false
}

func CreateGetFunctionsMsgFromJSON(jsonString string) (*GetFunctionsMsg, error) {
	var msg *GetFunctionsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
