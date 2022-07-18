package etcd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEtcdCluster(t *testing.T) {
	node1 := Node{Name: "etcd1", Host: "localhost", Port: 24100, PeerPort: 23100}
	node2 := Node{Name: "etcd2", Host: "localhost", Port: 24200, PeerPort: 23200}
	node3 := Node{Name: "etcd3", Host: "localhost", Port: 24300, PeerPort: 23300}
	node4 := Node{Name: "etcd4", Host: "localhost", Port: 24400, PeerPort: 23400}
	cluster := Cluster{}
	cluster.AddNode(node1)
	cluster.AddNode(node2)
	cluster.AddNode(node3)
	cluster.AddNode(node4)

	server1 := CreateEtcdServer(node1, cluster, ".")
	server2 := CreateEtcdServer(node2, cluster, ".")
	server3 := CreateEtcdServer(node3, cluster, ".")
	server4 := CreateEtcdServer(node4, cluster, ".")

	assert.Equal(t, server1.buildInitialClusterStr(), "etcd1=http://localhost:23100,etcd2=http://localhost:23200,etcd3=http://localhost:23300,etcd4=http://localhost:23400")

	server1.Start()
	server2.Start()
	server3.Start()
	server4.Start()

	server1.WaitToStart()
	server2.WaitToStart()
	server3.WaitToStart()
	server4.WaitToStart()

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

	server1.Stop()
	server2.Stop()
	server3.Stop()
	server4.Stop()

	server1.WaitToStop()
	server2.WaitToStop()
	server3.WaitToStop()
	server4.WaitToStop()

	os.RemoveAll(server1.StorageDir())
	os.RemoveAll(server2.StorageDir())
	os.RemoveAll(server3.StorageDir())
	os.RemoveAll(server4.StorageDir())
}
