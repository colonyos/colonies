package rpc

import (
	"encoding/json"
)

const ChannelReadPayloadType = "channelreadmsg"

type ChannelReadMsg struct {
	ProcessID string `json:"processid"`
	Name      string `json:"name"`
	AfterSeq  int64  `json:"afterseq"`
	Limit     int    `json:"limit"`
	MsgType   string `json:"msgtype"`
}

func CreateChannelReadMsg(processID string, name string, afterSeq int64, limit int) *ChannelReadMsg {
	msg := &ChannelReadMsg{}
	msg.ProcessID = processID
	msg.Name = name
	msg.AfterSeq = afterSeq
	msg.Limit = limit
	msg.MsgType = ChannelReadPayloadType

	return msg
}

func (msg *ChannelReadMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *ChannelReadMsg) Equals(msg2 *ChannelReadMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType &&
		msg.ProcessID == msg2.ProcessID &&
		msg.Name == msg2.Name &&
		msg.AfterSeq == msg2.AfterSeq &&
		msg.Limit == msg2.Limit {
		return true
	}

	return false
}

func CreateChannelReadMsgFromJSON(jsonString string) (*ChannelReadMsg, error) {
	var msg *ChannelReadMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
