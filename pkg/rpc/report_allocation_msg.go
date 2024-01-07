package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const ReportAllocationsPayloadType = "reportallocationmsg"

type ReportAllocationsMsg struct {
	ColonyName   string           `json:"colonyname"`
	ExecutorName string           `json:"executorname"`
	Allocations  core.Allocations `json:"allocations"`
	MsgType      string           `json:"msgtype"`
}

func CreateReportAllocationsMsg(colonyName string, executorName string, allocations core.Allocations) *ReportAllocationsMsg {
	msg := &ReportAllocationsMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.Allocations = allocations
	msg.MsgType = ReportAllocationsPayloadType

	return msg
}

func (msg *ReportAllocationsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ReportAllocationsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ReportAllocationsMsg) Equals(msg2 *ReportAllocationsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && core.IsProjectsEqual(msg.Allocations.Projects, msg2.Allocations.Projects) && msg.ExecutorName == msg2.ExecutorName && msg.ColonyName == msg2.ColonyName {
		return true
	}

	return false
}

func CreateReportAllocationsMsgFromJSON(jsonString string) (*ReportAllocationsMsg, error) {
	var msg *ReportAllocationsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
