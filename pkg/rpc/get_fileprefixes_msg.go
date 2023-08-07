package rpc

import (
	"encoding/json"
)

const GetFilePrefixesPayloadType = "getfileprefixesmsg"

type GetFilePrefixesMsg struct {
	MsgType  string `json:"msgtype"`
	ColonyID string `json:"colonyid"`
}

func CreateGetFilePrefixesMsg(colonyID string) *GetFilePrefixesMsg {
	msg := &GetFilePrefixesMsg{}
	msg.ColonyID = colonyID
	msg.MsgType = GetFilePrefixesPayloadType

	return msg
}

func (msg *GetFilePrefixesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetFilePrefixesMsg) Equals(msg2 *GetFilePrefixesMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID {
		return true
	}

	return false
}

func (msg *GetFilePrefixesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetFilePrefixesMsgFromJSON(jsonString string) (*GetFilePrefixesMsg, error) {
	var msg *GetFilePrefixesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
