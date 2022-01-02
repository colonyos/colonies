package rpc

import (
	"colonies/pkg/security/crypto"
	"encoding/base64"
	"encoding/json"
)

type RPCMsg struct {
	Method    string `json:"method"`
	Signature string `json:"signature"`
	Payload   string `json:"payload"`
	Err       bool   `json:"error"`
}

func CreateRPCMsg(method string, payload string, prvKey string) (*RPCMsg, error) {
	msg := &RPCMsg{}
	msg.Method = method
	msg.Payload = payload
	msg.Err = false
	msg.Payload = base64.StdEncoding.EncodeToString([]byte(payload))

	signature, err := crypto.CreateCrypto().GenerateSignature(msg.Payload, prvKey)
	if err != nil {
		return nil, err
	}

	msg.Signature = signature

	return msg, nil
}

func CreateRPCMsgFromJSON(jsonString string) (*RPCMsg, error) {
	var msg *RPCMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}

func (msg *RPCMsg) DecodePayload() string {
	jsonBytes, _ := base64.StdEncoding.DecodeString(msg.Payload)

	return string(jsonBytes)
}

func (msg *RPCMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *RPCMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
