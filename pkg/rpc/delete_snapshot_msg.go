package rpc

import (
	"encoding/json"
)

const DeleteSnapshotPayloadType = "deletesnapshotmsg"

type DeleteSnapshotMsg struct {
	ColonyID   string `json:"colonyid"`
	SnapshotID string `json:"snapshotid"`
	Name       string `json:"name"`
	MsgType    string `json:"msgtype"`
}

func CreateDeleteSnapshotMsg(colonyID string, snapshotID string, name string) *DeleteSnapshotMsg {
	msg := &DeleteSnapshotMsg{}
	msg.MsgType = DeleteSnapshotPayloadType
	msg.ColonyID = colonyID
	msg.SnapshotID = snapshotID
	msg.Name = name

	return msg
}

func (msg *DeleteSnapshotMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteSnapshotMsg) Equals(msg2 *DeleteSnapshotMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyID == msg2.ColonyID &&
		msg.Name == msg2.Name &&
		msg.SnapshotID == msg2.SnapshotID {
		return true
	}

	return false
}

func CreateDeleteSnapshotMsgFromJSON(jsonString string) (*DeleteSnapshotMsg, error) {
	var msg *DeleteSnapshotMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
