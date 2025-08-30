package rpc

import (
	"encoding/json"
)

const GetPauseStatusPayloadType = "getpausestatusmsg"

type GetPauseStatusMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
}

func CreateGetPauseStatusMsg(colonyName string) *GetPauseStatusMsg {
	msg := &GetPauseStatusMsg{}
	msg.MsgType = GetPauseStatusPayloadType
	msg.ColonyName = colonyName
	return msg
}

func (msg *GetPauseStatusMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetPauseStatusMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetPauseStatusMsg) Equals(msg2 *GetPauseStatusMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateGetPauseStatusMsgFromJSON(jsonString string) (*GetPauseStatusMsg, error) {
	var msg *GetPauseStatusMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}