package rpc

import "encoding/json"

const GetServiceHistoryPayloadType = "getservicehistorymsg"

type GetServiceHistoryMsg struct {
	ServiceID string `json:"serviceid"`
	Limit     int    `json:"limit,omitempty"`
	MsgType   string `json:"msgtype"`
}

func CreateGetServiceHistoryMsg(serviceID string, limit int) *GetServiceHistoryMsg {
	msg := &GetServiceHistoryMsg{}
	msg.ServiceID = serviceID
	msg.Limit = limit
	msg.MsgType = GetServiceHistoryPayloadType

	return msg
}

func CreateGetServiceHistoryMsgFromJSON(jsonString string) (*GetServiceHistoryMsg, error) {
	var msg GetServiceHistoryMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (msg *GetServiceHistoryMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
