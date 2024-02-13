package p2p

import "encoding/json"

type Node struct {
	Addr   []string
	HostID string
}

func CreateNode(hostID string, addr []string) Node {
	return Node{
		Addr:   addr,
		HostID: hostID,
	}
}

func (n *Node) String() string {
	str := "Node{" + n.HostID + ", ["
	for _, addr := range n.Addr {
		str += addr + ", "
	}

	if len(n.Addr) > 0 {
		str = str[:len(str)-2]
	}

	str += "]}"
	return str
}

func (n *Node) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(n)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func ConvertJSONToNode(jsonStr string) (Node, error) {
	var n Node
	err := json.Unmarshal([]byte(jsonStr), &n)
	if err != nil {
		return n, err
	}

	return n, nil
}
