package rpc

import (
	"encoding/json"
)

const ResetFunctionStatsPayloadType = "resetfunctionstatsmsg"

type ResetFunctionStatsMsg struct {
	ColonyName string `json:"colonyname"`
	MsgType    string `json:"msgtype"`
}

func CreateResetFunctionStatsMsg(colonyName string) *ResetFunctionStatsMsg {
	msg := &ResetFunctionStatsMsg{}
	msg.ColonyName = colonyName
	msg.MsgType = ResetFunctionStatsPayloadType

	return msg
}

func (msg *ResetFunctionStatsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ResetFunctionStatsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ResetFunctionStatsMsg) Equals(msg2 *ResetFunctionStatsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateResetFunctionStatsMsgFromJSON(jsonString string) (*ResetFunctionStatsMsg, error) {
	var msg *ResetFunctionStatsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
