package p2p

type Node struct {
	Addr   []string
	HostID string
}

func CreateNode(hostID string, addr []string) *Node {
	return &Node{
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
