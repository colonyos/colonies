package rpc

import (
	"encoding/json"
)

const SubscribeFilesPayloadType = "subscribefilesmsg"

// SubscribeFilesMsg is sent over a websocket to register a subscription for
// file events under a label prefix. The server fans out matching FileEvents
// to the subscriber until the websocket closes or Timeout (seconds) elapses.
//
// LabelPrefix matches a file event's Label when:
//   - LabelPrefix == "" (matches all labels in the colony), or
//   - event.Label == LabelPrefix (exact match), or
//   - event.Label starts with LabelPrefix + "/".
//
// Kinds is an OR-filter on the FileEventKind ints (1=added, 2=updated,
// 3=removed). An empty Kinds list means "all kinds".
type SubscribeFilesMsg struct {
	ColonyName  string `json:"colonyname"`
	LabelPrefix string `json:"labelprefix"`
	Kinds       []int  `json:"kinds"`
	Timeout     int    `json:"timeout"`
	MsgType     string `json:"msgtype"`
}

func CreateSubscribeFilesMsg(colonyName string, labelPrefix string, kinds []int, timeout int) *SubscribeFilesMsg {
	msg := &SubscribeFilesMsg{}
	msg.ColonyName = colonyName
	msg.LabelPrefix = labelPrefix
	msg.Kinds = kinds
	msg.Timeout = timeout
	msg.MsgType = SubscribeFilesPayloadType

	return msg
}

func (msg *SubscribeFilesMsg) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubscribeFilesMsg) ToJSONIndent() (string, error) {
	jsonBytes, err := json.MarshalIndent(msg, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (msg *SubscribeFilesMsg) Equals(msg2 *SubscribeFilesMsg) bool {
	if msg2 == nil {
		return false
	}

	if msg.ColonyName != msg2.ColonyName ||
		msg.MsgType != msg2.MsgType ||
		msg.LabelPrefix != msg2.LabelPrefix ||
		msg.Timeout != msg2.Timeout {
		return false
	}

	if len(msg.Kinds) != len(msg2.Kinds) {
		return false
	}
	for i := range msg.Kinds {
		if msg.Kinds[i] != msg2.Kinds[i] {
			return false
		}
	}
	return true
}

func CreateSubscribeFilesMsgFromJSON(jsonString string) (*SubscribeFilesMsg, error) {
	var msg *SubscribeFilesMsg

	err := json.Unmarshal([]byte(jsonString), &msg)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
