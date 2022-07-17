package server

import (
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

	assert.Equal(t, cluster.buildInitialClusterStr(), "etcd1=http://localhost:23100,etcd2=http://localhost:23200,etcd3=http://localhost:23300,etcd4=http://localhost:23400")

	server1 := CreateEtcdServer(node1, cluster)
	server2 := CreateEtcdServer(node2, cluster)
	server3 := CreateEtcdServer(node3, cluster)
	server4 := CreateEtcdServer(node4, cluster)

	ready1 := server1.Start()
	ready2 := server2.Start()
	ready3 := server3.Start()
	ready4 := server4.Start()

	<-ready1
	<-ready2
	<-ready3
	<-ready4

	leader := server1.Leader()
	assert.Equal(t, server2.Leader(), leader)
	assert.Equal(t, server3.Leader(), leader)
	assert.Equal(t, server4.Leader(), leader)
}
