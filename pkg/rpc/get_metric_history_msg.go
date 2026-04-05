package rpc

import (
	"encoding/json"
	"time"
)

const GetMetricHistoryPayloadType = "getmetrichistorymsg"

type GetMetricHistoryMsg struct {
	ColonyName   string    `json:"colonyname"`
	ExecutorName string    `json:"executorname"`
	Key          string    `json:"key"`
	Period       int       `json:"period"`
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	MsgType      string    `json:"msgtype"`
}

func CreateGetMetricHistoryMsg(colonyName string, executorName string, key string, period int, from time.Time, to time.Time) *GetMetricHistoryMsg {
	msg := &GetMetricHistoryMsg{}
	msg.ColonyName = colonyName
	msg.ExecutorName = executorName
	msg.Key = key
	msg.Period = period
	msg.From = from
	msg.To = to
	msg.MsgType = GetMetricHistoryPayloadType
	return msg
}

func (msg *GetMetricHistoryMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *GetMetricHistoryMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (msg *GetMetricHistoryMsg) Equals(msg2 *GetMetricHistoryMsg) bool {
	if msg2 == nil {
		return false
	}
	if msg.MsgType == msg2.MsgType &&
		msg.ColonyName == msg2.ColonyName &&
		msg.ExecutorName == msg2.ExecutorName &&
		msg.Key == msg2.Key &&
		msg.Period == msg2.Period &&
		msg.From.Equal(msg2.From) &&
		msg.To.Equal(msg2.To) {
		return true
	}
	return false
}

func CreateGetMetricHistoryMsgFromJSON(jsonString string) (*GetMetricHistoryMsg, error) {
	var msg *GetMetricHistoryMsg
	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}
	return msg, nil
}
