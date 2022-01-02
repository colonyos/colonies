package rpc

import (
	"encoding/json"
)

const GetRuntimesMsgType = "getruntimes"

type GetRuntimesMsg struct {
	ColonyID string `json:"colonyid"`
}

func CreateGetRuntimesMsg(colonyID string) *GetRuntimesMsg {
	msg := &GetRuntimesMsg{}
	msg.ColonyID = colonyID

	return msg
}

func (msg *GetRuntimesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetRuntimesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateGetRuntimesMsgFromJSON(jsonString string) (*GetRuntimesMsg, error) {
	var msg *GetRuntimesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
