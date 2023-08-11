package rpc

import (
	"encoding/json"
)

const GetSnapshotsPayloadType = "getsnapshotsmsg"

type GetSnapshotsMsg struct {
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
}

func CreateGetSnapshotsMsg(colonyID string) *GetSnapshotsMsg {
	msg := &GetSnapshotsMsg{}
	msg.MsgType = GetSnapshotsPayloadType
	msg.ColonyID = colonyID

	return msg
}

func (msg *GetSnapshotsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetSnapshotsMsg) Equals(msg2 *GetSnapshotsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyID == msg2.ColonyID {
		return true
	}

	return false
}

func CreateGetSnapshotsMsgFromJSON(jsonString string) (*GetSnapshotsMsg, error) {
	var msg *GetSnapshotsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
