package rpc

import (
	"encoding/json"
)

const GetLocationsPayloadType = "getlocationsmsg"

type GetLocationsMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
}

func CreateGetLocationsMsg(colonyName string) *GetLocationsMsg {
	msg := &GetLocationsMsg{}
	msg.MsgType = GetLocationsPayloadType
	msg.ColonyName = colonyName

	return msg
}

func (msg *GetLocationsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetLocationsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetLocationsMsg) Equals(msg2 *GetLocationsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateGetLocationsMsgFromJSON(jsonString string) (*GetLocationsMsg, error) {
	var msg *GetLocationsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
