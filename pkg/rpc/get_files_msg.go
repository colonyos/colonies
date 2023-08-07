package rpc

import (
	"encoding/json"
)

const GetFilesPayloadType = "getfilesmsg"

type GetFilesMsg struct {
	Prefix   string `json:"prefix"`
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
}

func CreateGetFilesMsg(prefix string, colonyID string) *GetFilesMsg {
	msg := &GetFilesMsg{}
	msg.Prefix = prefix
	msg.ColonyID = colonyID
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

	if msg.MsgType == msg2.MsgType && msg.Prefix == msg2.Prefix && msg.ColonyID == msg2.ColonyID {
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
