package rpc

import (
	"encoding/json"
)

const GetFileLabelsPayloadType = "getfilelabelsmsg"

type GetFileLabelsMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
	Name       string `json:"name"`
	Exact      bool   `json:"exact"`
}

func CreateGetFileLabelsMsg(colonyName string, name string, exact bool) *GetFileLabelsMsg {
	msg := &GetFileLabelsMsg{}
	msg.ColonyName = colonyName
	msg.Name = name
	msg.Exact = exact
	msg.MsgType = GetFileLabelsPayloadType

	return msg
}

func CreateGetAllFileLabelsMsg(colonyName string) *GetFileLabelsMsg {
	msg := &GetFileLabelsMsg{}
	msg.ColonyName = colonyName
	msg.Name = ""
	msg.MsgType = GetFileLabelsPayloadType

	return msg
}

func (msg *GetFileLabelsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetFileLabelsMsg) Equals(msg2 *GetFileLabelsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName && msg.Name == msg2.Name && msg.Exact == msg2.Exact {
		return true
	}

	return false
}

func (msg *GetFileLabelsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetFileLabelsMsgFromJSON(jsonString string) (*GetFileLabelsMsg, error) {
	var msg *GetFileLabelsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
