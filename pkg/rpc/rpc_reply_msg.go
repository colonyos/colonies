package rpc

import (
	"encoding/base64"
	"encoding/json"
)

const ErrorPayloadType = "error"

type RPCReplyMsg struct {
	PayloadType string `json:"payloadtype"`
	Payload     string `json:"payload"`
	Error       bool   `json:"error"`
}

func CreateRPCReplyMsg(payloadType string, payload string) (*RPCReplyMsg, error) {
	msg := &RPCReplyMsg{}
	msg.PayloadType = payloadType
	msg.Payload = payload
	msg.Payload = base64.StdEncoding.EncodeToString([]byte(payload))
	msg.Error = false

	return msg, nil
}

func CreateRPCErrorReplyMsg(payloadType string, payload string) (*RPCReplyMsg, error) {
	msg := &RPCReplyMsg{}
	msg.PayloadType = payloadType
	msg.Payload = payload
	msg.Payload = base64.StdEncoding.EncodeToString([]byte(payload))
	msg.Error = true

	return msg, nil
}

func CreateRPCReplyMsgFromJSON(jsonString string) (*RPCReplyMsg, error) {
	var msg *RPCReplyMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}

func (msg *RPCReplyMsg) DecodePayload() string {
	jsonBytes, _ := base64.StdEncoding.DecodeString(msg.Payload)

	return string(jsonBytes)
}

func (msg *RPCReplyMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RPCReplyMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
