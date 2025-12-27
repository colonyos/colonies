package rpc

import (
	"encoding/json"
)

const UpdateBlueprintStatusPayloadType = "updateblueprintstatusmsg"

type UpdateBlueprintStatusMsg struct {
	ColonyName    string                 `json:"colonyname"`
	BlueprintName string                 `json:"blueprintname"`
	Status        map[string]interface{} `json:"status"`
	MsgType       string                 `json:"msgtype"`
}

func CreateUpdateBlueprintStatusMsg(colonyName, blueprintName string, status map[string]interface{}) *UpdateBlueprintStatusMsg {
	msg := &UpdateBlueprintStatusMsg{}
	msg.ColonyName = colonyName
	msg.BlueprintName = blueprintName
	msg.Status = status
	msg.MsgType = UpdateBlueprintStatusPayloadType

	return msg
}

func (msg *UpdateBlueprintStatusMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *UpdateBlueprintStatusMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *UpdateBlueprintStatusMsg) Equals(msg2 *UpdateBlueprintStatusMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType != msg2.MsgType {
		return false
	}

	if msg.ColonyName != msg2.ColonyName {
		return false
	}

	if msg.BlueprintName != msg2.BlueprintName {
		return false
	}

	return true
}

func CreateUpdateBlueprintStatusMsgFromJSON(jsonString string) (*UpdateBlueprintStatusMsg, error) {
	var msg *UpdateBlueprintStatusMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
