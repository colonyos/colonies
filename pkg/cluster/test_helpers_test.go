package cluster

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
)

// testClusterNodes returns n cluster nodes with kernel-assigned free ports so
// tests never collide on fixed port numbers.
func testClusterNodes(t *testing.T, names ...string) []Node {
	t.Helper()
	ports, err := utils.FreePorts(4 * len(names))
	if err != nil {
		t.Fatalf("failed to allocate ports: %v", err)
	}

	nodes := make([]Node, len(names))
	for i, name := range names {
		nodes[i] = Node{
			Name:           name,
			Host:           "localhost",
			EtcdClientPort: ports[4*i],
			EtcdPeerPort:   ports[4*i+1],
			RelayPort:      ports[4*i+2],
			APIPort:        ports[4*i+3],
		}
	}
	return nodes
}
