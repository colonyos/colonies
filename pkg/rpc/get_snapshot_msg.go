package rpc

import (
	"encoding/json"
)

const GetSnapshotPayloadType = "getsnapshotmsg"

type GetSnapshotMsg struct {
	ColonyName string `json:"colonyname"`
	SnapshotID string `json:"snapshotid"`
	Name       string `json:"name"`
	MsgType    string `json:"msgtype"`
}

func CreateGetSnapshotMsg(colonyName string, snapshotID string, name string) *GetSnapshotMsg {
	msg := &GetSnapshotMsg{}
	msg.MsgType = GetSnapshotPayloadType
	msg.ColonyName = colonyName
	msg.SnapshotID = snapshotID
	msg.Name = name

	return msg
}

func (msg *GetSnapshotMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetSnapshotMsg) Equals(msg2 *GetSnapshotMsg) bool {
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

func CreateGetSnapshotMsgFromJSON(jsonString string) (*GetSnapshotMsg, error) {
	var msg *GetSnapshotMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
