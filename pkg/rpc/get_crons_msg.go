package rpc

import (
	"encoding/json"
)

const GetCronsPayloadType = "getcronsmsg"

type GetCronsMsg struct {
	ColonyName string `json:"colonyname"`
	Count      int    `json:"count"`
	MsgType    string `json:"msgtype"`
}

func CreateGetCronsMsg(colonyID string, count int) *GetCronsMsg {
	msg := &GetCronsMsg{}
	msg.ColonyName = colonyID
	msg.Count = count
	msg.MsgType = GetCronsPayloadType

	return msg
}

func (msg *GetCronsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetCronsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetCronsMsg) Equals(msg2 *GetCronsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.Count == msg2.Count {
		return true
	}

	return false
}

func CreateGetCronsMsgFromJSON(jsonString string) (*GetCronsMsg, error) {
	var msg *GetCronsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
