package rpc

import (
	"encoding/json"
)

const SubscribeChannelPayloadType = "subscribechannelmsg"

type SubscribeChannelMsg struct {
	ProcessID string `json:"processid"`
	Name      string `json:"name"`
	AfterSeq  int64  `json:"afterseq"`
	Timeout   int    `json:"timeout"`
	MsgType   string `json:"msgtype"`
}

func CreateSubscribeChannelMsg(processID string, name string, afterSeq int64, timeout int) *SubscribeChannelMsg {
	msg := &SubscribeChannelMsg{}
	msg.ProcessID = processID
	msg.Name = name
	msg.AfterSeq = afterSeq
	msg.Timeout = timeout
	msg.MsgType = SubscribeChannelPayloadType

	return msg
}

func (msg *SubscribeChannelMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubscribeChannelMsg) Equals(msg2 *SubscribeChannelMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ProcessID == msg2.ProcessID &&
		msg.Name == msg2.Name &&
		msg.AfterSeq == msg2.AfterSeq &&
		msg.Timeout == msg2.Timeout {
		return true
	}

	return false
}

func CreateSubscribeChannelMsgFromJSON(jsonString string) (*SubscribeChannelMsg, error) {
	var msg *SubscribeChannelMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
