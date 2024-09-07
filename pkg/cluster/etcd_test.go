package cluster

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEtcdCluster(t *testing.T) {
	node1 := &Node{Name: "etcd1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := &Node{Name: "etcd2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := &Node{Name: "etcd3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}
	node4 := &Node{Name: "etcd4", Host: "localhost", EtcdClientPort: 24400, EtcdPeerPort: 23400, RelayPort: 25400, APIPort: 26400}

	config := EmptyConfig()
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)
	config.AddNode(node4)

	server1 := CreateEtcdServer(node1, config, ".")
	server2 := CreateEtcdServer(node2, config, ".")
	server3 := CreateEtcdServer(node3, config, ".")
	server4 := CreateEtcdServer(node4, config, ".")

	// The order of the nodes in the initial cluster may be come in any order
	// "etcd4=http://localhost:23400,etcd1=http://localhost:23100,etcd2=http://localhost:23200,etcd3=http://localhost:23300"
	// "etcd1=http://localhost:23100,etcd2=http://localhost:23200,etcd3=http://localhost:23300,etcd4=http://localhost:23400"

	// Split the string into an array of strings
	// Then sort the array
	// Then join the array back into a string
	clusterStr := server1.buildInitialClusterStr()
	s := strings.Split(clusterStr, ",")
	sort.Strings(s)
	clusterStr = strings.Join(s, ",")

	assert.Equal(t, clusterStr, "etcd1=http://localhost:23100,etcd2=http://localhost:23200,etcd3=http://localhost:23300,etcd4=http://localhost:23400")

	server1.Start()
	server2.Start()
	server3.Start()
	server4.Start()

	server1.BlockUntilReady()
	server2.BlockUntilReady()
	server3.BlockUntilReady()
	server4.BlockUntilReady()

	leader := server1.Leader()
	assert.Equal(t, server2.Leader(), leader)
	assert.Equal(t, server3.Leader(), leader)
	assert.Equal(t, server4.Leader(), leader)

	nodes1 := server1.Members()
	assert.Len(t, nodes1, 4)
	nodes2 := server1.Members()
	assert.Len(t, nodes2, 4)
	nodes3 := server1.Members()
	assert.Len(t, nodes3, 4)
	nodes4 := server1.Members()
	assert.Len(t, nodes4, 4)

	currentCluster := server4.CurrentCluster()
	assert.Len(t, currentCluster.Nodes, 4)

	server1.Stop()
	server2.Stop()
	server3.Stop()
	server4.Stop()

	server1.BlockUntilStopped()
	server2.BlockUntilStopped()
	server3.BlockUntilStopped()
	server4.BlockUntilStopped()

	os.RemoveAll(server1.StorageDir())
	os.RemoveAll(server2.StorageDir())
	os.RemoveAll(server3.StorageDir())
	os.RemoveAll(server4.StorageDir())
}
