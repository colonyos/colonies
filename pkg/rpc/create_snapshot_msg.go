package rpc

import (
	"encoding/json"
)

const CreateSnapshotPayloadType = "createsnapshotmsg"

type CreateSnapshotMsg struct {
	ColonyName string `json:"colonyname"`
	Label      string `json:"label"`
	Name       string `json:"name"`
	MsgType    string `json:"msgtype"`
}

func CreateCreateSnapshotMsg(colonyID string, label string, name string) *CreateSnapshotMsg {
	msg := &CreateSnapshotMsg{}
	msg.MsgType = CreateSnapshotPayloadType
	msg.ColonyName = colonyID
	msg.Label = label
	msg.Name = name

	return msg
}

func (msg *CreateSnapshotMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *CreateSnapshotMsg) Equals(msg2 *CreateSnapshotMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.Label == msg2.Label &&
		msg.Name == msg2.Name {
		return true
	}

	return false
}

func CreateCreateSnapshotMsgFromJSON(jsonString string) (*CreateSnapshotMsg, error) {
	var msg *CreateSnapshotMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
