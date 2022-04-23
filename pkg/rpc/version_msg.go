package rpc

import (
	"encoding/json"
)

const VersionPayloadType = "versionmsg"

type VersionMsg struct {
	BuildVersion string `json:"buildversion"`
	BuildTime    string `json:"buildtime"`
	MsgType      string `json:"msgtype"`
}

func CreateVersionMsg(buildVersion string, buildTime string) *VersionMsg {
	msg := &VersionMsg{}
	msg.BuildVersion = buildVersion
	msg.BuildTime = buildTime
	msg.MsgType = VersionPayloadType

	return msg
}

func (msg *VersionMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *VersionMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *VersionMsg) Equals(msg2 *VersionMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.BuildVersion == msg2.BuildVersion &&
		msg.BuildTime == msg2.BuildTime {
		return true
	}

	return false
}

func CreateVersionMsgFromJSON(jsonString string) (*VersionMsg, error) {
	var msg *VersionMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
