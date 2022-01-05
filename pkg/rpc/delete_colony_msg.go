package rpc

import (
	"encoding/json"
)

const DeleteColonyPayloadType = "deletecolonymsg"

type DeleteColonyMsg struct {
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
}

func CreateDeleteColonyMsg(colonyID string) *DeleteColonyMsg {
	msg := &DeleteColonyMsg{}
	msg.ColonyID = colonyID
	msg.MsgType = DeleteColonyPayloadType

	return msg
}

func (msg *DeleteColonyMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *DeleteColonyMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateDeleteColonyMsgFromJSON(jsonString string) (*DeleteColonyMsg, error) {
	var msg *DeleteColonyMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
