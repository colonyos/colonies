package rpc

import (
	"encoding/json"
)

const DeleteAllProcessGraphsPayloadType = "deleteallprocessgraphsmsg"

type DeleteAllProcessGraphsMsg struct {
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
}

func CreateDeleteAllProcessGraphsMsg(colonyID string) *DeleteAllProcessGraphsMsg {
	msg := &DeleteAllProcessGraphsMsg{}
	msg.ColonyID = colonyID
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

	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID {
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
