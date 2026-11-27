package cluster

// Coordinator abstracts cluster coordination: membership, leader election and
// cross-server coordination state (the assignments-paused flags).
//
// Two implementations exist:
//   - *EtcdServer: an embedded etcd cluster; used whenever clustering is enabled.
//   - *LocalCoordinator: single-node semantics requiring no etcd. It is only
//     valid when the Colonies server runs with clustering disabled.
type Coordinator interface {
	// Lifecycle
	Start()
	WaitToStart()
	Stop()
	WaitToStop()
	// StorageDir returns the on-disk data directory, or "" if the coordinator
	// keeps no persistent state. Callers must not remove empty directories.
	StorageDir() string

	// Cluster membership
	Leader() string
	Members() []Node
	CurrentCluster() Config

	// Cross-server coordination state
	PauseColonyAssignments(colonyName string) error
	ResumeColonyAssignments(colonyName string) error
	AreColonyAssignmentsPaused(colonyName string) (bool, error)
}

// The embedded etcd server is one Coordinator implementation.
var _ Coordinator = (*EtcdServer)(nil)
