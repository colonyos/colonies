package etcd

import "encoding/json"

type Node struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	ClientPort int    `json:"port"`     // Etcd default 2379
	PeerPort   int    `json:"peerport"` // Etcd default 2380
}

func (node *Node) Equals(node2 *Node) bool {
	if node.Name != node2.Name {
		return false
	}
	if node.Host != node2.Host {
		return false
	}
	if node.ClientPort != node2.ClientPort {
		return false
	}
	if node.PeerPort != node2.PeerPort {
		return false
	}

	return true
}

type Cluster struct {
	Nodes  []Node `json:"nodes"`
	Leader Node   `json:"leader"`
}

func (cluster *Cluster) AddNode(node Node) {
	cluster.Nodes = append(cluster.Nodes, node)
}

func ConvertJSONToCluster(jsonString string) (*Cluster, error) {
	var cluster *Cluster
	err := json.Unmarshal([]byte(jsonString), &cluster)
	if err != nil {
		return cluster, err
	}

	return cluster, nil
}

func (cluster *Cluster) Equals(cluster2 *Cluster) bool {
	if cluster2 == nil {
		return false
	}

	if !cluster.Leader.Equals(&cluster2.Leader) {
		return false
	}

	counter := 0
	for _, node1 := range cluster.Nodes {
		for _, node2 := range cluster2.Nodes {
			if node1.Equals(&node2) {
				counter++
			}
		}
	}

	if counter == len(cluster.Nodes) && counter == len(cluster.Nodes) {
		return true
	}

	return false
}

func (cluster *Cluster) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(cluster, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
