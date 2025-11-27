package rpc

import (
	"encoding/json"
)

const ChannelAppendPayloadType = "channelappendmsg"

type ChannelAppendMsg struct {
	ProcessID string `json:"processid"`
	Name      string `json:"name"`
	Sequence  int64  `json:"sequence"`           // Client-assigned sequence number
	InReplyTo int64  `json:"inreplyto,omitempty"` // References sequence from other sender
	Payload   []byte `json:"payload"`
	MsgType   string `json:"msgtype"`
}

func CreateChannelAppendMsg(processID string, name string, sequence int64, inReplyTo int64, payload []byte) *ChannelAppendMsg {
	msg := &ChannelAppendMsg{}
	msg.ProcessID = processID
	msg.Name = name
	msg.Sequence = sequence
	msg.InReplyTo = inReplyTo
	msg.Payload = payload
	msg.MsgType = ChannelAppendPayloadType

	return msg
}

func (msg *ChannelAppendMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ChannelAppendMsg) Equals(msg2 *ChannelAppendMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ProcessID == msg2.ProcessID &&
		msg.Name == msg2.Name {
		return true
	}

	return false
}

func CreateChannelAppendMsgFromJSON(jsonString string) (*ChannelAppendMsg, error) {
	var msg *ChannelAppendMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
