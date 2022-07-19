package etcd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCluster(t *testing.T) {
	node1 := Node{Name: "etcd1", Host: "localhost", ClientPort: 24100, PeerPort: 23100}
	node2 := Node{Name: "etcd2", Host: "localhost", ClientPort: 24200, PeerPort: 23200}
	node3 := Node{Name: "etcd3", Host: "localhost", ClientPort: 24300, PeerPort: 23300}
	node4 := Node{Name: "etcd4", Host: "localhost", ClientPort: 24400, PeerPort: 23400}
	cluster := Cluster{}
	cluster.AddNode(node1)
	cluster.AddNode(node2)
	cluster.AddNode(node3)
	cluster.AddNode(node4)
	cluster.Leader = node2

	jsonStr, err := cluster.ToJSON()

	cluster2, err := ConvertJSONToCluster(jsonStr)
	assert.Nil(t, err)
	assert.True(t, cluster.Equals(cluster2))
	cluster.Leader = node1
	assert.False(t, cluster.Equals(cluster2))
}
