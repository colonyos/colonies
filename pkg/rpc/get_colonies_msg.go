package rpc

import (
	"encoding/json"
)

const GetColoniesMsgType = "getcolonies"

type GetColoniesMsg struct {
}

func CreateGetColoniesMsg() *GetColoniesMsg {
	msg := &GetColoniesMsg{}

	return msg
}

func (msg *GetColoniesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetColoniesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetColoniesMsgFromJSON(jsonString string) (*GetColoniesMsg, error) {
	var msg *GetColoniesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
