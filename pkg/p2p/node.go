package p2p

import "encoding/json"

type Node struct {
	Name   string `json:"name"`
	Addr   string `json:"addr"`
	HostID string `json:"hostID"`
}

func CreateNode(name string, hostID string, addr string) Node {
	return Node{
		Name:   name,
		Addr:   addr,
		HostID: hostID,
	}
}

func (n *Node) String() string {
	return "Node{" + n.Name + ":" + n.HostID + ", [" + string(n.Addr) + "]}"
}

func (n *Node) ToJSON() (string, error) {
	jsonBytes, err := json.Marshal(n)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (n *Node) Equals(other Node) bool {
	if n.Name != other.Name {
		return false
	}

	if n.HostID != other.HostID {
		return false
	}

	if n.Addr != other.Addr {
		return false
	}

	return true
}

func ConvertJSONToNode(jsonStr string) (Node, error) {
	var n Node
	err := json.Unmarshal([]byte(jsonStr), &n)
	if err != nil {
		return n, err
	}

	return n, nil
}
