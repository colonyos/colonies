package rpc

import "encoding/json"

const GetBlueprintHistoryPayloadType = "getblueprinthistorymsg"

type GetBlueprintHistoryMsg struct {
	BlueprintID string `json:"blueprintid"`
	Limit     int    `json:"limit,omitempty"`
	MsgType   string `json:"msgtype"`
}

func CreateGetBlueprintHistoryMsg(blueprintID string, limit int) *GetBlueprintHistoryMsg {
	msg := &GetBlueprintHistoryMsg{}
	msg.BlueprintID = blueprintID
	msg.Limit = limit
	msg.MsgType = GetBlueprintHistoryPayloadType

	return msg
}

func CreateGetBlueprintHistoryMsgFromJSON(jsonString string) (*GetBlueprintHistoryMsg, error) {
	var msg GetBlueprintHistoryMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func (msg *GetBlueprintHistoryMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
