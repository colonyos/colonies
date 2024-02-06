package network

const (
	MSG_PING_REQ int = iota
	MSG_PING_RESP
	MSG_FIND_CONTACTS_REQ
	MSG_FIND_CONTACTS_RESP
)

type Message struct {
	From    string
	To      string
	Type    int
	Payload []byte
	ID      string
}
