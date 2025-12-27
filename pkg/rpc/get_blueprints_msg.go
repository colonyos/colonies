package rpc

import (
	"encoding/json"
)

const GetBlueprintsPayloadType = "getblueprintsmsg"

type GetBlueprintsMsg struct {
	Namespace    string `json:"namespace"`
	Kind         string `json:"kind"`
	LocationName string `json:"locationname"`
	MsgType      string `json:"msgtype"`
}

func CreateGetBlueprintsMsg(namespace, kind, locationName string) *GetBlueprintsMsg {
	msg := &GetBlueprintsMsg{}
	msg.Namespace = namespace
	msg.Kind = kind
	msg.LocationName = locationName
	msg.MsgType = GetBlueprintsPayloadType

	return msg
}

func (msg *GetBlueprintsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetBlueprintsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetBlueprintsMsg) Equals(msg2 *GetBlueprintsMsg) bool {
	if msg2 == nil {
		return false
	}

	return msg.MsgType == msg2.MsgType &&
		msg.Namespace == msg2.Namespace &&
		msg.Kind == msg2.Kind &&
		msg.LocationName == msg2.LocationName
}

func CreateGetBlueprintsMsgFromJSON(jsonString string) (*GetBlueprintsMsg, error) {
	var msg *GetBlueprintsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
