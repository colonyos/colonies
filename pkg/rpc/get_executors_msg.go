package rpc

import (
	"encoding/json"
)

const GetExecutorsPayloadType = "getexecutorsmsg"

type GetExecutorsMsg struct {
	ColonyName           string `json:"colonyname"`
	IncludeUnregistered  bool   `json:"includeunregistered"`
	MsgType              string `json:"msgtype"`
}

func CreateGetExecutorsMsg(colonyName string, includeUnregistered bool) *GetExecutorsMsg {
	msg := &GetExecutorsMsg{}
	msg.ColonyName = colonyName
	msg.IncludeUnregistered = includeUnregistered
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

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName && msg.IncludeUnregistered == msg2.IncludeUnregistered {
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
