package rpc

import (
	"encoding/json"
)

const SearchLogsPayloadType = "searchlogsmsg"

type SearchLogsMsg struct {
	ColonyName string `json:"colonyname"`
	Text       string `json:"text"`
	Days       int    `json:"days"`
	Count      int    `json:"count"`
	MsgType    string `json:"msgtype"`
}

func CreateSearchLogsMsg(colonyName string, text string, days int, count int) *SearchLogsMsg {
	msg := &SearchLogsMsg{}
	msg.ColonyName = colonyName
	msg.Text = text
	msg.Days = days
	msg.Count = count
	msg.MsgType = SearchLogsPayloadType

	return msg
}

func (msg *SearchLogsMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SearchLogsMsg) Equals(msg2 *SearchLogsMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.MsgType == msg2.MsgType && msg.ColonyName == msg2.ColonyName && msg.Text == msg2.Text && msg.Days == msg2.Days && msg.Count == msg2.Count {
		return true
	}

	return false
}

func (msg *SearchLogsMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func CreateSearchLogsMsgFromJSON(jsonString string) (*SearchLogsMsg, error) {
	var msg *SearchLogsMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
