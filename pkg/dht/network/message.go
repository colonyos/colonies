package network

const (
	MSG_PING_REQ int = iota
	MSG_PING_RESP
	MSG_FIND_CONTACTS_REQ
	MSG_FIND_CONTACTS_RESP
	MSG_PUT_REQ
	MSG_PUT_RESP
	MSG_GET_REQ
	MSG_GET_RESP
)

type Message struct {
	ID      string
	From    string
	To      string
	Type    int
	Payload []byte
}
