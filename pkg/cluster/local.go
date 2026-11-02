package cluster

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// LocalCoordinator is the single-node Coordinator. It requires no etcd and
// provides in-process semantics:
//   - the local node is always the leader;
//   - the cluster is a singleton;
//   - assignments-paused flags live in memory, so unlike the etcd Coordinator
//     a paused colony resumes assignments when the server restarts. This is
//     documented in docs/Configuration.md.
//
// The in-memory coordination state is only correct for a single server
// process. Deployments with more than one Colonies server sharing a database
// must use clustering (the etcd Coordinator). The server cannot detect other
// servers on the same database, so the operator must ensure this; the CLI only
// rejects --initial-cluster when clustering is disabled.
type LocalCoordinator struct {
	thisNode      Node
	clusterConfig Config

	pauseMutex     sync.Mutex
	pausedColonies map[string]bool
}

// CreateLocalCoordinator creates the single-node Coordinator for the given node.
func CreateLocalCoordinator(thisNode Node) *LocalCoordinator {
	config := Config{}
	config.AddNode(thisNode)
	config.Leader = thisNode

	log.WithFields(log.Fields{
		"Node":    thisNode.Name,
		"APIPort": thisNode.APIPort,
	}).Info("Clustering disabled, using local (single-node) coordinator")

	return &LocalCoordinator{
		thisNode:       thisNode,
		clusterConfig:  config,
		pausedColonies: make(map[string]bool),
	}
}

// Lifecycle: a LocalCoordinator holds no resources and no persistent state.
func (c *LocalCoordinator) Start()             {}
func (c *LocalCoordinator) WaitToStart()       {}
func (c *LocalCoordinator) Stop()              {}
func (c *LocalCoordinator) WaitToStop()        {}
func (c *LocalCoordinator) StorageDir() string { return "" }

// Cluster membership: the local node is always the leader.
func (c *LocalCoordinator) Leader() string {
	return c.thisNode.Name
}

func (c *LocalCoordinator) Members() []Node {
	return []Node{c.thisNode}
}

func (c *LocalCoordinator) CurrentCluster() Config {
	return c.clusterConfig
}

// Coordination state: in-memory assignments-paused flags.
func (c *LocalCoordinator) PauseColonyAssignments(colonyName string) error {
	c.pauseMutex.Lock()
	defer c.pauseMutex.Unlock()
	c.pausedColonies[colonyName] = true

	log.WithFields(log.Fields{"Colony": colonyName}).Info("Colony process assignments have been paused")
	return nil
}

func (c *LocalCoordinator) ResumeColonyAssignments(colonyName string) error {
	c.pauseMutex.Lock()
	defer c.pauseMutex.Unlock()
	delete(c.pausedColonies, colonyName)

	log.WithFields(log.Fields{"Colony": colonyName}).Info("Colony process assignments have been resumed")
	return nil
}

func (c *LocalCoordinator) AreColonyAssignmentsPaused(colonyName string) (bool, error) {
	c.pauseMutex.Lock()
	defer c.pauseMutex.Unlock()
	return c.pausedColonies[colonyName], nil
}

var _ Coordinator = (*LocalCoordinator)(nil)
