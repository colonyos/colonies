package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateCluster(t *testing.T) {
	node1 := &Node{Name: "etcd1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := &Node{Name: "etcd2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := &Node{Name: "etcd3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}
	node4 := &Node{Name: "etcd4", Host: "localhost", EtcdClientPort: 24400, EtcdPeerPort: 23400, RelayPort: 25400, APIPort: 26400}

	config := EmptyConfig()
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)
	config.AddNode(node4)
	config.Leader = node2

	jsonStr, err := config.ToJSON()

	config2, err := ConvertJSONToConfig(jsonStr)
	assert.Nil(t, err)
	assert.True(t, config.Equals(config2))
	config.Leader = node1
	assert.False(t, config.Equals(config2))
}
