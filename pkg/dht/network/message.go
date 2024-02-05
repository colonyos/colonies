package network

const (
	MSG_PING_REQ int = iota
	MSG_PING_RESP
	MSG_FIND_NODE_REQ
	MSG_FIND_NODE_RESP
)

type Message struct {
	FromAddr string
	ToAddr   string
	Type     int
	Payload  []byte
}
