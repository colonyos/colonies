package rpc

import (
	"encoding/json"
)

const GetColoniesMsgType = "GetColonies"

type GetColoniesMsg struct {
	RPC          RPC    `json:"rpc"`
	RootPassword string `json:"rootpassword"`
}

func CreateGetColoniesMsg(rootPassword string) *GetColoniesMsg {
	msg := &GetColoniesMsg{}
	msg.RPC.Method = GetColoniesMsgType
	msg.RootPassword = rootPassword

	return msg
}

func (msg *GetColoniesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
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
