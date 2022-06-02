package rpc

import (
	"encoding/json"
)

const GetStatisiticsPayloadType = "getstatisticsmsg"

type GetStatisticsMsg struct {
	MsgType string `json:"msgtype"`
}

func CreateGetStatisticsMsg() *GetStatisticsMsg {
	msg := &GetStatisticsMsg{}
	msg.MsgType = GetStatisiticsPayloadType

	return msg
}

func (msg *GetStatisticsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetStatisticsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *GetStatisticsMsg) Equals(msg2 *GetStatisticsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType {
		return true
	}

	return false
}

func CreateGetStatisticsMsgFromJSON(jsonString string) (*GetStatisticsMsg, error) {
	var msg *GetStatisticsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
