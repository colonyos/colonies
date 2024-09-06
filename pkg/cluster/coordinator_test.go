package cluster

import (
	"io"
	"testing"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCoordinatorGenNodeList(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	log.SetLevel(log.DebugLevel)

	node1 := &Node{Name: "replica1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := &Node{Name: "replica2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := &Node{Name: "replica3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := EmptyConfig()
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	clusterManager1 := CreateClusterManager(node1, config, ".")
	clusterManager2 := CreateClusterManager(node2, config, ".")
	clusterManager3 := CreateClusterManager(node3, config, ".")

	clusterManager1.BlockUntilReady()
	clusterManager2.BlockUntilReady()
	clusterManager3.BlockUntilReady()

	coordinator := clusterManager1.Coordinator()

	doneChan := make(chan bool, 2)

	go func() {
		inProgress := coordinator.genNodeList()
		doneChan <- inProgress
	}()

	go func() {
		inProgress := coordinator.genNodeList()
		doneChan <- inProgress
	}()

	progress1 := <-doneChan
	progress2 := <-doneChan

	assert.True(t, progress1 != progress2)

	coordinator.genNodeList()

	nodeList := coordinator.GetNodeList()
	assert.Equal(t, 3, len(nodeList))

	for _, node := range nodeList {
		assert.True(t, node == "replica1" || node == "replica2" || node == "replica3")
	}

	clusterManager1.Shutdown()
	clusterManager2.Shutdown()
	clusterManager3.Shutdown()
}

func TestCoordinatorGetNodeListFromLeader(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	log.SetLevel(log.DebugLevel)

	node1 := &Node{Name: "replica1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := &Node{Name: "replica2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := &Node{Name: "replica3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := EmptyConfig()
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	clusterManager1 := CreateClusterManager(node1, config, ".")
	clusterManager2 := CreateClusterManager(node2, config, ".")
	clusterManager3 := CreateClusterManager(node3, config, ".")

	clusterManager1.BlockUntilReady()
	clusterManager2.BlockUntilReady()
	clusterManager3.BlockUntilReady()

	nodeList1, err := clusterManager1.Coordinator().getNodeListFromLeader(false)
	assert.Nil(t, err)

	nodeList2, err := clusterManager2.Coordinator().getNodeListFromLeader(false)
	assert.Nil(t, err)

	nodeList3, err := clusterManager3.Coordinator().getNodeListFromLeader(false)
	assert.Nil(t, err)

	assert.Equal(t, 3, len(nodeList1))
	assert.Equal(t, 3, len(nodeList2))
	assert.Equal(t, 3, len(nodeList3))

	for _, node := range nodeList1 {
		assert.True(t, node == "replica1" || node == "replica2" || node == "replica3")
	}

	for _, node := range nodeList2 {
		assert.True(t, node == "replica1" || node == "replica2" || node == "replica3")
	}

	for _, node := range nodeList3 {
		assert.True(t, node == "replica1" || node == "replica2" || node == "replica3")
	}

	clusterManager1.Shutdown()
	clusterManager2.Shutdown()
	clusterManager3.Shutdown()
}

func TestCoordinatorVerifyNodeList(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	log.SetLevel(log.DebugLevel)

	node1 := &Node{Name: "replica1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := &Node{Name: "replica2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := &Node{Name: "replica3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := EmptyConfig()
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	clusterManager1 := CreateClusterManager(node1, config, ".")
	clusterManager2 := CreateClusterManager(node2, config, ".")
	clusterManager3 := CreateClusterManager(node3, config, ".")

	clusterManager1.BlockUntilReady()
	clusterManager2.BlockUntilReady()
	clusterManager3.BlockUntilReady()

	nodeList1, err := clusterManager1.Coordinator().getNodeListFromLeader(false)
	assert.Nil(t, err)

	nodeList2, err := clusterManager2.Coordinator().getNodeListFromLeader(false)
	assert.Nil(t, err)

	nodeList3, err := clusterManager3.Coordinator().getNodeListFromLeader(false)
	assert.Nil(t, err)

	assert.Equal(t, 3, len(nodeList1))
	assert.Equal(t, 3, len(nodeList2))
	assert.Equal(t, 3, len(nodeList3))

	for _, node := range nodeList1 {
		assert.True(t, node == "replica1" || node == "replica2" || node == "replica3")
	}

	for _, node := range nodeList2 {
		assert.True(t, node == "replica1" || node == "replica2" || node == "replica3")
	}

	for _, node := range nodeList3 {
		assert.True(t, node == "replica1" || node == "replica2" || node == "replica3")
	}

	ok, err := clusterManager1.Coordinator().verifyNodeList(nodeList1)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = clusterManager2.Coordinator().verifyNodeList(nodeList1)
	assert.Nil(t, err)
	assert.True(t, ok)

	ok, err = clusterManager3.Coordinator().verifyNodeList(nodeList1)
	assert.Nil(t, err)
	assert.True(t, ok)

	missingNodeList := []string{"replica1", "replica2"}
	ok, err = clusterManager1.Coordinator().verifyNodeList(missingNodeList)
	assert.Nil(t, err)
	assert.False(t, ok)

	missingNodeList = []string{"replica1"}
	ok, err = clusterManager2.Coordinator().verifyNodeList(missingNodeList)
	assert.Nil(t, err)
	assert.False(t, ok)

	missingNodeList = []string{"replica1", "replica3"}
	ok, err = clusterManager3.Coordinator().verifyNodeList(missingNodeList)
	assert.Nil(t, err)
	assert.False(t, ok)

	clusterManager1.Shutdown()
	clusterManager2.Shutdown()
	clusterManager3.Shutdown()
}

func TestFindNode(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	log.SetLevel(log.DebugLevel)

	node1 := &Node{Name: "replica1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := &Node{Name: "replica2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := &Node{Name: "replica3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := EmptyConfig()
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	clusterManager1 := CreateClusterManager(node1, config, ".")
	clusterManager2 := CreateClusterManager(node2, config, ".")
	clusterManager3 := CreateClusterManager(node3, config, ".")

	clusterManager1.BlockUntilReady()
	clusterManager2.BlockUntilReady()
	clusterManager3.BlockUntilReady()

	id := core.GenerateRandomID()
	n1, err := clusterManager1.Coordinator().FindNode(id, false)
	assert.Nil(t, err)

	n2, err := clusterManager1.Coordinator().FindNode(id, false)
	assert.Nil(t, err)

	n3, err := clusterManager1.Coordinator().FindNode(id, false)
	assert.Nil(t, err)

	assert.True(t, n1 == n2)
	assert.True(t, n1 == n3)

	clusterManager1.Shutdown()
	clusterManager2.Shutdown()
	clusterManager3.Shutdown()
}
