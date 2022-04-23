package rpc

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/colonyos/colonies/pkg/security/crypto"
)

type RPCMsg struct {
	Signature   string `json:"signature"`
	PayloadType string `json:"payloadtype"`
	Payload     string `json:"payload"`
}

func CreateRPCMsg(payloadType string, payload string, prvKey string) (*RPCMsg, error) {
	msg := &RPCMsg{}
	msg.PayloadType = payloadType
	msg.Payload = base64.StdEncoding.EncodeToString([]byte(payload))

	signature, err := crypto.CreateCrypto().GenerateSignature(msg.Payload, prvKey)
	if err != nil {
		return nil, errors.New("Failed to generate signature")
	}

	msg.Signature = signature

	return msg, nil
}

func CreateInsecureRPCMsg(payloadType string, payload string) (*RPCMsg, error) {
	msg := &RPCMsg{}
	msg.PayloadType = payloadType
	msg.Payload = base64.StdEncoding.EncodeToString([]byte(payload))

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

func (msg *RPCMsg) Equals(msg2 *RPCMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.Signature == msg2.Signature &&
		msg.PayloadType == msg2.PayloadType &&
		msg.Payload == msg2.Payload {
		return true
	}

	return false
}
