package rpc

import (
	"encoding/json"
)

const GetSnapshotPayloadType = "getsnapshotmsg"

type GetSnapshotMsg struct {
	ColonyID   string `json:"colonyid"`
	SnapshotID string `json:"snapshotid"`
	MsgType    string `json:"msgtype"`
}

func CreateGetSnapshotMsg(colonyID string, snapshotID string) *GetSnapshotMsg {
	msg := &GetSnapshotMsg{}
	msg.MsgType = GetSnapshotPayloadType
	msg.ColonyID = colonyID
	msg.SnapshotID = snapshotID

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
		msg.ColonyID == msg2.ColonyID &&
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
