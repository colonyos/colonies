package rpc

import (
	"encoding/json"
)

const DeleteFilePayloadType = "deletefilemsg"

type DeleteFileMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
	FileID     string `json:"fileid"`
	Label      string `json:"label"`
	Name       string `json:"name"`
}

func CreateDeleteFileMsg(colonyID string, fileID string, label string, name string) *DeleteFileMsg {
	msg := &DeleteFileMsg{}
	msg.ColonyName = colonyID
	msg.FileID = fileID
	msg.Label = label
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
		msg.ColonyName == msg2.ColonyName &&
		msg.FileID == msg2.FileID &&
		msg.Label == msg2.Label &&
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
