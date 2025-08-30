package rpc

import (
	"encoding/json"
)

const PauseAssignmentsPayloadType = "pauseassignmentsmsg"

type PauseAssignmentsMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
}

func CreatePauseAssignmentsMsg(colonyName string) *PauseAssignmentsMsg {
	msg := &PauseAssignmentsMsg{}
	msg.MsgType = PauseAssignmentsPayloadType
	msg.ColonyName = colonyName
	return msg
}

func (msg *PauseAssignmentsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *PauseAssignmentsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *PauseAssignmentsMsg) Equals(msg2 *PauseAssignmentsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreatePauseAssignmentsMsgFromJSON(jsonString string) (*PauseAssignmentsMsg, error) {
	var msg *PauseAssignmentsMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}