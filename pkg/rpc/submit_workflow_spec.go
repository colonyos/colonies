package rpc

import (
	"encoding/json"

	"github.com/colonyos/colonies/pkg/core"
)

const SubmitWorkflowSpecPayloadType = "submitworkflowspecmsg"

type SubmitWorkflowSpecMsg struct {
	WorkflowSpec *core.WorkflowSpec `json:"spec"`
	MsgType      string             `json:"msgtype"`
}

func CreateSubmitWorkflowSpecMsg(workflowSpec *core.WorkflowSpec) *SubmitWorkflowSpecMsg {
	msg := &SubmitWorkflowSpecMsg{}
	msg.WorkflowSpec = workflowSpec
	msg.MsgType = SubmitWorkflowSpecPayloadType

	return msg
}

func (msg *SubmitWorkflowSpecMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubmitWorkflowSpecMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubmitWorkflowSpecMsg) Equals(msg2 *SubmitWorkflowSpecMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.WorkflowSpec.Equals(msg2.WorkflowSpec) {
		return true
	}

	return false
}

func CreateSubmitWorkflowSpecMsgFromJSON(jsonString string) (*SubmitWorkflowSpecMsg, error) {
	var msg *SubmitWorkflowSpecMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
