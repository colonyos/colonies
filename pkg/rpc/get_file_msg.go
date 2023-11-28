package rpc

import (
	"encoding/json"
)

const GetFilePayloadType = "getfilemsg"

type GetFileMsg struct {
	ColonyName string `json:"colonyname"`
	FileID     string `json:"fileid"`
	Label      string `json:"label"`
	Name       string `json:"name"`
	Latest     bool   `json:"latest"`
	MsgType    string `json:"msgtype"`
}

func CreateGetFileMsg(colonyID string, fileID string, label string, name string, latest bool) *GetFileMsg {
	msg := &GetFileMsg{}
	msg.ColonyName = colonyID
	msg.FileID = fileID
	msg.Label = label
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
		msg.ColonyName == msg2.ColonyName &&
		msg.FileID == msg2.FileID &&
		msg.Label == msg2.Label &&
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
