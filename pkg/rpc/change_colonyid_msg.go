package rpc

import (
	"encoding/json"
)

const ChangeColonyIDPayloadType = "changecolonyidmsg"

type ChangeColonyIDMsg struct {
	ColonyName string `json:"colonyname"`
	ColonyID   string `json:"colonyid"`
	MsgType    string `json:"msgtype"`
}

func CreateChangeColonyIDMsg(colonyName, colonyID string) *ChangeColonyIDMsg {
	msg := &ChangeColonyIDMsg{}
	msg.ColonyName = colonyName
	msg.ColonyID = colonyID
	msg.MsgType = ChangeColonyIDPayloadType

	return msg
}

func (msg *ChangeColonyIDMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ChangeColonyIDMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ChangeColonyIDMsg) Equals(msg2 *ChangeColonyIDMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateChangeColonyIDMsgFromJSON(jsonString string) (*ChangeColonyIDMsg, error) {
	var msg *ChangeColonyIDMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
