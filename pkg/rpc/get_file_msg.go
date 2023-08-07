package rpc

import (
	"encoding/json"
)

const GetFilePayloadType = "getfilemsg"

type GetFileMsg struct {
	ColonyID string `json:"colonyid"`
	FileID   string `json:"fileid"`
	Prefix   string `json:"prefix"`
	Name     string `json:"name"`
	Latest   bool   `json:"latest"`
	MsgType  string `json:"msgtype"`
}

func CreateGetFileMsg(colonyID string, fileID string, prefix string, name string, latest bool) *GetFileMsg {
	msg := &GetFileMsg{}
	msg.ColonyID = colonyID
	msg.FileID = fileID
	msg.Prefix = prefix
	msg.Name = name
	msg.Latest = latest
	msg.MsgType = GetFilePayloadType

	return msg
}

func (msg *GetFileMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetFileMsg) Equals(msg2 *GetFileMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyID == msg2.ColonyID &&
		msg.FileID == msg2.FileID &&
		msg.Prefix == msg2.Prefix &&
		msg.Name == msg2.Name &&
		msg.Latest == msg2.Latest {
		return true
	}

	return false
}

func (msg *GetFileMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetFileMsgFromJSON(jsonString string) (*GetFileMsg, error) {
	var msg *GetFileMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
