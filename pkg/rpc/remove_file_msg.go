package rpc

import (
	"encoding/json"
)

const RemoveFilePayloadType = "removefilemsg"

type RemoveFileMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
	FileID     string `json:"fileid"`
	Label      string `json:"label"`
	Name       string `json:"name"`
}

func CreateRemoveFileMsg(colonyName string, fileID string, label string, name string) *RemoveFileMsg {
	msg := &RemoveFileMsg{}
	msg.ColonyName = colonyName
	msg.FileID = fileID
	msg.Label = label
	msg.Name = name
	msg.MsgType = RemoveFilePayloadType

	return msg
}

func (msg *RemoveFileMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveFileMsg) Equals(msg2 *RemoveFileMsg) bool {
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

func (msg *RemoveFileMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateRemoveFileMsgFromJSON(jsonString string) (*RemoveFileMsg, error) {
	var msg *RemoveFileMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
