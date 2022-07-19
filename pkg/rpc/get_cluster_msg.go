package rpc

import (
	"encoding/json"
)

const GetClusterPayloadType = "getclustermsg"

type GetClusterMsg struct {
	MsgType string `json:"msgtype"`
}

func CreateGetClusterMsg() *GetClusterMsg {
	msg := &GetClusterMsg{}
	msg.MsgType = GetClusterPayloadType

	return msg
}

func (msg *GetClusterMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetClusterMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetClusterMsg) Equals(msg2 *GetClusterMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType {
		return true
	}

	return false
}

func CreateGetClusterMsgFromJSON(jsonString string) (*GetClusterMsg, error) {
	var msg *GetClusterMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
