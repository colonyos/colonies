package rpc

import (
	"colonies/pkg/core"
	"encoding/json"
)

const AddColonyMsgType = "AddColony"

type AddColonyMsg struct {
	RPC    RPC          `json:"rpc"`
	Colony *core.Colony `json:"colony"`
}

func CreateAddColonyMsg(colony *core.Colony) *AddColonyMsg {
	msg := &AddColonyMsg{}
	msg.RPC.Method = AddColonyMsgType
	msg.Colony = colony

	return msg
}

func (msg *AddColonyMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateAddColonyMsgFromJSON(jsonString string) (*AddColonyMsg, error) {
	var msg *AddColonyMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
