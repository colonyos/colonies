package rpc

import (
	"encoding/json"
)

const GetFileLabelsPayloadType = "getfilelabelsmsg"

type GetFileLabelsMsg struct {
	MsgType  string `json:"msgtype"`
	ColonyID string `json:"colonyid"`
}

func CreateGetFileLabelsMsg(colonyID string) *GetFileLabelsMsg {
	msg := &GetFileLabelsMsg{}
	msg.ColonyID = colonyID
	msg.MsgType = GetFileLabelsPayloadType

	return msg
}

func (msg *GetFileLabelsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetFileLabelsMsg) Equals(msg2 *GetFileLabelsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID {
		return true
	}

	return false
}

func (msg *GetFileLabelsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetFileLabelsMsgFromJSON(jsonString string) (*GetFileLabelsMsg, error) {
	var msg *GetFileLabelsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
