package rpc

import (
	"encoding/json"
)

const PauseStatusReplyPayloadType = "pausestatusreplymsg"

type PauseStatusReplyMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
	IsPaused   bool   `json:"ispaused"`
}

func CreatePauseStatusReplyMsg(colonyName string, isPaused bool) *PauseStatusReplyMsg {
	msg := &PauseStatusReplyMsg{}
	msg.MsgType = PauseStatusReplyPayloadType
	msg.ColonyName = colonyName
	msg.IsPaused = isPaused
	return msg
}

func (msg *PauseStatusReplyMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *PauseStatusReplyMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *PauseStatusReplyMsg) Equals(msg2 *PauseStatusReplyMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName && msg.IsPaused == msg2.IsPaused {
		return true
	}

	return false
}

func CreatePauseStatusReplyMsgFromJSON(jsonString string) (*PauseStatusReplyMsg, error) {
	var msg *PauseStatusReplyMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}