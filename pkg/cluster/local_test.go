package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalCoordinatorIsAlwaysLeader(t *testing.T) {
	node := Node{Name: "single", Host: "localhost", EtcdClientPort: 2379, EtcdPeerPort: 2380, RelayPort: 2381, APIPort: 8080}
	coordinator := CreateLocalCoordinator(node)

	assert.Equal(t, "single", Coordinator(coordinator).Leader())
	assert.Equal(t, []Node{node}, coordinator.Members())
}

func TestLocalCoordinatorClusterConfig(t *testing.T) {
	node := Node{Name: "single", Host: "localhost", APIPort: 8080}
	coordinator := CreateLocalCoordinator(node)

	config := Coordinator(coordinator).CurrentCluster()
	assert.Len(t, config.Nodes, 1)
	assert.Equal(t, "single", config.Leader.Name)
}

func TestLocalCoordinatorPause(t *testing.T) {
	coordinator := CreateLocalCoordinator(Node{Name: "single", Host: "localhost", APIPort: 8080})

	paused, err := coordinator.AreColonyAssignmentsPaused("colony1")
	assert.Nil(t, err)
	assert.False(t, paused)

	err = coordinator.PauseColonyAssignments("colony1")
	assert.Nil(t, err)

	paused, err = coordinator.AreColonyAssignmentsPaused("colony1")
	assert.Nil(t, err)
	assert.True(t, paused)

	err = coordinator.ResumeColonyAssignments("colony1")
	assert.Nil(t, err)

	paused, err = coordinator.AreColonyAssignmentsPaused("colony1")
	assert.Nil(t, err)
	assert.False(t, paused)
}

func TestLocalCoordinatorLifecycle(t *testing.T) {
	coordinator := CreateLocalCoordinator(Node{Name: "single", Host: "localhost", APIPort: 8080})

	// Lifecycle calls must be no-ops and safe to call in any order
	coordinator.Start()
	coordinator.WaitToStart()
	coordinator.Stop()
	coordinator.WaitToStop()

	// No persistent state: StorageDir is empty so Stop paths never RemoveAll("")
	assert.Equal(t, "", coordinator.StorageDir())
}
