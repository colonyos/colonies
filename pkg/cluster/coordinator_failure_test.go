package cluster

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestCoordinatorGenNodeListFailure(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	log.SetLevel(log.DebugLevel)

	node1 := Node{Name: "replica1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := Node{Name: "replica2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := Node{Name: "replica3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	clusterManager1 := CreateClusterManager(node1, config, ".")
	clusterManager2 := CreateClusterManager(node2, config, ".")
	clusterManager3 := CreateClusterManager(node3, config, ".")

	clusterManager1.BlockUntilReady()
	clusterManager2.BlockUntilReady()
	clusterManager3.BlockUntilReady()

	coordinator1 := clusterManager1.Coordinator()
	fmt.Println(coordinator1)
	coordinator1.EnableFailures(time.Duration(10) * time.Second) // Will delay ping response in 10 seconds

	coordinator3 := clusterManager3.Coordinator()

	coordinator3.genNodeList()
	nodeList := coordinator3.GetNodeList()
	assert.Equal(t, 2, len(nodeList))
	for _, node := range nodeList {
		assert.True(t, node == "replica2" || node == "replica3")
	}

	clusterManager1.Shutdown()
	clusterManager2.Shutdown()
	clusterManager3.Shutdown()
}
