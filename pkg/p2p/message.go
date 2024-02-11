package p2p

type Message struct {
	ID      string
	From    Node
	To      Node
	Type    int
	Payload []byte
}
