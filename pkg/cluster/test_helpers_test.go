package cluster

import (
	"testing"

	"github.com/colonyos/colonies/pkg/utils"
)

// testClusterNodes returns n cluster nodes with kernel-assigned free ports so
// tests never collide on fixed port numbers. The ports stay reserved until the
// returned release function is called, which tests should do right before the
// first component binds. Unreleased reservations are released when the test
// ends.
func testClusterNodes(t *testing.T, names ...string) ([]Node, func()) {
	t.Helper()
	reserved, err := utils.ReservePorts(4 * len(names))
	if err != nil {
		t.Fatalf("failed to reserve ports: %v", err)
	}
	release := func() { utils.ReleasePorts(reserved) }
	t.Cleanup(release)

	nodes := make([]Node, len(names))
	for i, name := range names {
		nodes[i] = Node{
			Name:           name,
			Host:           "localhost",
			EtcdClientPort: reserved[4*i].Port(),
			EtcdPeerPort:   reserved[4*i+1].Port(),
			RelayPort:      reserved[4*i+2].Port(),
			APIPort:        reserved[4*i+3].Port(),
		}
	}
	return nodes, release
}
