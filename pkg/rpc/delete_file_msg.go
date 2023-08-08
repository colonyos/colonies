package rpc

import (
	"encoding/json"
)

const DeleteFilePayloadType = "deletefilemsg"

type DeleteFileMsg struct {
	MsgType  string `json:"msgtype"`
	ColonyID string `json:"colonyid"`
	FileID   string `json:"fileid"`
	Prefix   string `json:"prefix"`
	Name     string `json:"name"`
}

func CreateDeleteFileMsg(colonyID string, fileID string, prefix string, name string) *DeleteFileMsg {
	msg := &DeleteFileMsg{}
	msg.ColonyID = colonyID
	msg.FileID = fileID
	msg.Prefix = prefix
	msg.Name = name
	msg.MsgType = DeleteFilePayloadType

	return msg
}

func (msg *DeleteFileMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteFileMsg) Equals(msg2 *DeleteFileMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyID == msg2.ColonyID &&
		msg.FileID == msg2.FileID &&
		msg.Prefix == msg2.Prefix &&
		msg.Name == msg2.Name {
		return true
	}

	return false
}

func (msg *DeleteFileMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateDeleteFileMsgFromJSON(jsonString string) (*DeleteFileMsg, error) {
	var msg *DeleteFileMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
