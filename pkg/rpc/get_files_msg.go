package rpc

import (
	"encoding/json"
)

const GetFilesPayloadType = "getfilesmsg"

type GetFilesMsg struct {
	Label    string `json:"label"`
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
}

func CreateGetFilesMsg(colonyID string, label string) *GetFilesMsg {
	msg := &GetFilesMsg{}
	msg.ColonyID = colonyID
	msg.Label = label
	msg.MsgType = GetFilesPayloadType

	return msg
}

func (msg *GetFilesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetFilesMsg) Equals(msg2 *GetFilesMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID && msg.Label == msg2.Label {
		return true
	}

	return false
}

func (msg *GetFilesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetFilesMsgFromJSON(jsonString string) (*GetFilesMsg, error) {
	var msg *GetFilesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
