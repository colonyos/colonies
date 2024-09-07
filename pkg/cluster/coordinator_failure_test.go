package cluster

import (
	"io"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func findMissingElement(original, subset []string) string {
	elementCount := make(map[string]int)

	for _, v := range original {
		elementCount[v]++
	}

	for _, v := range subset {
		elementCount[v]--
	}

	for key, count := range elementCount {
		if count > 0 {
			return key
		}
	}

	return ""
}

func TestCoordinatorGenNodeListFailure(t *testing.T) {
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

	// Find out who the coordinator is the leader
	leaderName := clusterManager1.Coordinator().LeaderName()

	var leaderCoordinator *Coordinator

	if leaderName == "replica1" {
		leaderCoordinator = clusterManager1.Coordinator()
	} else if leaderName == "replica2" {
		leaderCoordinator = clusterManager2.Coordinator()
	} else if leaderName == "replica3" {
		leaderCoordinator = clusterManager3.Coordinator()
	}

	assert.NotNil(t, leaderCoordinator)

	var failedCoordinator *Coordinator
	if leaderName == "replica1" {
		failedCoordinator = clusterManager2.Coordinator()
	} else if leaderName == "replica2" {
		failedCoordinator = clusterManager1.Coordinator()
	} else if leaderName == "replica3" {
		failedCoordinator = clusterManager1.Coordinator()
	}
	failedName := failedCoordinator.Name()

	assert.NotNil(t, failedCoordinator)

	otherName := findMissingElement([]string{"replica1", "replica2", "replica3"}, []string{leaderName, failedName})

	// Find out who the other coordinator is
	var otherCoordinator *Coordinator
	if otherName == "replica1" {
		otherCoordinator = clusterManager1.Coordinator()
	} else if otherName == "replica2" {
		otherCoordinator = clusterManager2.Coordinator()
	} else if otherName == "replica3" {
		otherCoordinator = clusterManager3.Coordinator()
	}

	assert.NotNil(t, otherCoordinator)

	failedCoordinator.EnableFailures(time.Duration(PING_RESPONSE_TIMEOUT+1) * time.Second) // Will delay ping response in 10 seconds
	leaderCoordinator.genNodeList()

	nodeList := leaderCoordinator.GetNodeList()
	assert.Equal(t, 2, len(nodeList))
	for _, node := range nodeList {
		assert.True(t, node == leaderCoordinator.Name() || node == otherCoordinator.Name())
	}

	failedCoordinator.DisableFailures()

	// Wait for the coordinator to retry
	time.Sleep(time.Duration(NODE_LIST_RETRY_DELAY+1) * time.Second)

	// Full node list should now be generated
	nodeList = leaderCoordinator.GetNodeList()
	assert.Equal(t, 3, len(nodeList))
	for _, node := range nodeList {
		assert.True(t, node == "replica1" || node == "replica2" || node == "replica3")
	}

	clusterManager1.Shutdown()
	clusterManager2.Shutdown()
	clusterManager3.Shutdown()
}
