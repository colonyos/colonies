package rpc

import (
	"encoding/json"
)

const DeleteAllProcessesPayloadType = "deleteallprocessesmsg"

type DeleteAllProcessesMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
	State      int    `json:"state"`
}

func CreateDeleteAllProcessesMsg(colonyID string) *DeleteAllProcessesMsg {
	msg := &DeleteAllProcessesMsg{}
	msg.ColonyName = colonyID
	msg.MsgType = DeleteAllProcessesPayloadType

	return msg
}

func (msg *DeleteAllProcessesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteAllProcessesMsg) Equals(msg2 *DeleteAllProcessesMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName && msg.State == msg2.State {
		return true
	}

	return false
}

func (msg *DeleteAllProcessesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateDeleteAllProcessesMsgFromJSON(jsonString string) (*DeleteAllProcessesMsg, error) {
	var msg *DeleteAllProcessesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
