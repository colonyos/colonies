package rpc

import (
	"encoding/json"
)

const GetColonyStatisticsPayloadType = "getcolonystatsmsg"

type GetColonyStatisticsMsg struct {
	ColonyID string `json:"colonyid"`
	MsgType  string `json:"msgtype"`
}

func CreateGetColonyStatisticsMsg(colonyID string) *GetColonyStatisticsMsg {
	msg := &GetColonyStatisticsMsg{}
	msg.ColonyID = colonyID
	msg.MsgType = GetColonyStatisticsPayloadType

	return msg
}

func (msg *GetColonyStatisticsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetColonyStatisticsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetColonyStatisticsMsg) Equals(msg2 *GetColonyStatisticsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyID == msg2.ColonyID {
		return true
	}

	return false
}

func CreateGetColonyStatisticsMsgFromJSON(jsonString string) (*GetColonyStatisticsMsg, error) {
	var msg *GetColonyStatisticsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
