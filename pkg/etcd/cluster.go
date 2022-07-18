package etcd

type Node struct {
	Name     string
	Host     string
	Port     int // Etcd efault typically 2379
	PeerPort int // Etcd efault typically 2380
}

type Cluster struct {
	Nodes []Node
}

func (cluster *Cluster) AddNode(node Node) {
	cluster.Nodes = append(cluster.Nodes, node)
}
