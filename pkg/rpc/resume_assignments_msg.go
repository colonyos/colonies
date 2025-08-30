package rpc

import (
	"encoding/json"
)

const ResumeAssignmentsPayloadType = "resumeassignmentsmsg"

type ResumeAssignmentsMsg struct {
	MsgType    string `json:"msgtype"`
	ColonyName string `json:"colonyname"`
}

func CreateResumeAssignmentsMsg(colonyName string) *ResumeAssignmentsMsg {
	msg := &ResumeAssignmentsMsg{}
	msg.MsgType = ResumeAssignmentsPayloadType
	msg.ColonyName = colonyName
	return msg
}

func (msg *ResumeAssignmentsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ResumeAssignmentsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ResumeAssignmentsMsg) Equals(msg2 *ResumeAssignmentsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateResumeAssignmentsMsgFromJSON(jsonString string) (*ResumeAssignmentsMsg, error) {
	var msg *ResumeAssignmentsMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}