package rpc

import (
	"encoding/json"
)

const DeleteAllProcessGraphsPayloadType = "deleteallprocessgraphsmsg"

type DeleteAllProcessGraphsMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
	State      int    `json:"state"`
}

func CreateDeleteAllProcessGraphsMsg(colonyID string) *DeleteAllProcessGraphsMsg {
	msg := &DeleteAllProcessGraphsMsg{}
	msg.ColonyName = colonyID
	msg.MsgType = DeleteAllProcessGraphsPayloadType

	return msg
}

func (msg *DeleteAllProcessGraphsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteAllProcessGraphsMsg) Equals(msg2 *DeleteAllProcessGraphsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName && msg.State == msg2.State {
		return true
	}

	return false
}

func (msg *DeleteAllProcessGraphsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateDeleteAllProcessGraphsMsgFromJSON(jsonString string) (*DeleteAllProcessGraphsMsg, error) {
	var msg *DeleteAllProcessGraphsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
