package rpc

import (
	"encoding/json"
)

const RemoveAllSnapshotsPayloadType = "removeallsnapshotmsg"

type RemoveAllSnapshotsMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
}

func CreateRemoveAllSnapshotsMsg(colonyName string) *RemoveAllSnapshotsMsg {
	msg := &RemoveAllSnapshotsMsg{}
	msg.MsgType = RemoveAllSnapshotsPayloadType
	msg.ColonyName = colonyName

	return msg
}

func (msg *RemoveAllSnapshotsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RemoveAllSnapshotsMsg) Equals(msg2 *RemoveAllSnapshotsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateRemoveAllSnapshotsMsgFromJSON(jsonString string) (*RemoveAllSnapshotsMsg, error) {
	var msg *RemoveAllSnapshotsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
