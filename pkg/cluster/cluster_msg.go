package cluster

import (
	"github.com/vmihailenco/msgpack/v5"
)

const (
	PingRequest = iota
	PingResponse
	NodeListRequest
	NodeListResponse
	VerifyNodeListRequest
	VerifyNodeListResponse
	FindNodeRequest
	FindNodeResponse
)

type ClusterMsg struct {
	MsgType         int
	ID              string
	Originator      string
	Recipient       string
	NodeList        []string
	NodeListHash    string
	FailureDetected bool
}

func (m *ClusterMsg) Serialize() ([]byte, error) {
	return msgpack.Marshal(&m)
}

func DeserializeClusterMsg(data []byte) (*ClusterMsg, error) {
	var m ClusterMsg
	err := msgpack.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (m *ClusterMsg) Equals(other *ClusterMsg) bool {
	return m.MsgType == other.MsgType &&
		m.ID == other.ID &&
		m.Originator == other.Originator &&
		m.Recipient == other.Recipient &&
		m.FailureDetected == other.FailureDetected &&
		m.NodeListHash == other.NodeListHash
}
