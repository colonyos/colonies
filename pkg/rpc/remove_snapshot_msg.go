package rpc

import (
	"encoding/json"
)

const RemoveSnapshotPayloadType = "removesnapshotmsg"

type RemoveSnapshotMsg struct {
	ColonyName string `json:"colonyname"`
	SnapshotID string `json:"snapshotid"`
	Name       string `json:"name"`
	MsgType    string `json:"msgtype"`
}

func CreateRemoveSnapshotMsg(colonyName string, snapshotID string, name string) *RemoveSnapshotMsg {
	msg := &RemoveSnapshotMsg{}
	msg.MsgType = RemoveSnapshotPayloadType
	msg.ColonyName = colonyName
	msg.SnapshotID = snapshotID
	msg.Name = name

	return msg
}

func (msg *RemoveSnapshotMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveSnapshotMsg) Equals(msg2 *RemoveSnapshotMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.Name == msg2.Name &&
		msg.SnapshotID == msg2.SnapshotID {
		return true
	}

	return false
}

func CreateRemoveSnapshotMsgFromJSON(jsonString string) (*RemoveSnapshotMsg, error) {
	var msg *RemoveSnapshotMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
