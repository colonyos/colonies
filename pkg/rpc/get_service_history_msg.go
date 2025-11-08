package rpc

import "encoding/json"

const GetResourceHistoryPayloadType = "getresourcehistorymsg"

type GetResourceHistoryMsg struct {
	ResourceID string `json:"resourceid"`
	Limit      int    `json:"limit,omitempty"`
	MsgType    string `json:"msgtype"`
}

func CreateGetResourceHistoryMsg(resourceID string, limit int) *GetResourceHistoryMsg {
	msg := &GetResourceHistoryMsg{}
	msg.ResourceID = resourceID
	msg.Limit = limit
	msg.MsgType = GetResourceHistoryPayloadType

	return msg
}

func CreateGetResourceHistoryMsgFromJSON(jsonString string) (*GetResourceHistoryMsg, error) {
	var msg GetResourceHistoryMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (msg *GetResourceHistoryMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
