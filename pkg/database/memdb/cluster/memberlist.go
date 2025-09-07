package cluster

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"
)

// ClusterMember represents a member of the VelocityDB cluster
type ClusterMember struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Addr     string `json:"addr"`
	Port     uint16 `json:"port"`
	RaftAddr string `json:"raft_addr"`
	State    string `json:"state"` // alive, suspect, dead
	Meta     map[string]string `json:"meta,omitempty"`
}

// ClusterManager manages cluster membership using Memberlist
type ClusterManager struct {
	memberlist *memberlist.Memberlist
	config     *ClusterConfig
	eventCh    chan memberlist.NodeEvent
	members    map[string]*ClusterMember
	mu         sync.RWMutex
	localMember *ClusterMember
}

// ClusterConfig holds cluster configuration
type ClusterConfig struct {
	NodeName     string
	BindAddr     string
	BindPort     int
	RaftAddr     string
	SeedNodes    []string
	SecretKey    []byte
	LogOutput    bool
}

// NodeDelegate implements memberlist.Delegate
type NodeDelegate struct {
	manager *ClusterManager
}

// NewClusterManager creates a new cluster manager
func NewClusterManager(config *ClusterConfig) (*ClusterManager, error) {
	// Create memberlist config
	mlConfig := memberlist.DefaultWANConfig()
	mlConfig.Name = config.NodeName
	mlConfig.BindAddr = config.BindAddr
	mlConfig.BindPort = config.BindPort
	mlConfig.SecretKey = config.SecretKey
	
	if !config.LogOutput {
		mlConfig.LogOutput = nil
	}

	// Create cluster manager
	manager := &ClusterManager{
		config:  config,
		eventCh: make(chan memberlist.NodeEvent, 256),
		members: make(map[string]*ClusterMember),
		localMember: &ClusterMember{
			ID:       config.NodeName,
			Name:     config.NodeName,
			Addr:     config.BindAddr,
			Port:     uint16(config.BindPort),
			RaftAddr: config.RaftAddr,
			State:    "alive",
		},
	}

	// Set up delegate
	delegate := &NodeDelegate{manager: manager}
	mlConfig.Delegate = delegate
	mlConfig.Events = &EventDelegate{manager: manager}

	// Create memberlist
	ml, err := memberlist.Create(mlConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create memberlist: %w", err)
	}

	manager.memberlist = ml

	// Add local member
	manager.members[config.NodeName] = manager.localMember

	return manager, nil
}

// Join joins the cluster by contacting seed nodes
func (cm *ClusterManager) Join(seedNodes []string) error {
	if len(seedNodes) == 0 {
		seedNodes = cm.config.SeedNodes
	}

	if len(seedNodes) == 0 {
		// No seed nodes, we're the first node
		return nil
	}

	_, err := cm.memberlist.Join(seedNodes)
	if err != nil {
		return fmt.Errorf("failed to join cluster: %w", err)
	}

	return nil
}

// Leave gracefully leaves the cluster
func (cm *ClusterManager) Leave(timeout time.Duration) error {
	return cm.memberlist.Leave(timeout)
}

// GetMembers returns all cluster members
func (cm *ClusterManager) GetMembers() []*ClusterMember {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	members := make([]*ClusterMember, 0, len(cm.members))
	for _, member := range cm.members {
		members = append(members, member)
	}

	return members
}

// GetMember returns a specific cluster member
func (cm *ClusterManager) GetMember(name string) (*ClusterMember, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	member, exists := cm.members[name]
	return member, exists
}

// GetAliveMembers returns only alive cluster members
func (cm *ClusterManager) GetAliveMembers() []*ClusterMember {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var alive []*ClusterMember
	for _, member := range cm.members {
		if member.State == "alive" {
			alive = append(alive, member)
		}
	}

	return alive
}

// LocalMember returns the local cluster member
func (cm *ClusterManager) LocalMember() *ClusterMember {
	return cm.localMember
}

// NumMembers returns the number of members in the cluster
func (cm *ClusterManager) NumMembers() int {
	return cm.memberlist.NumMembers()
}

// NumAliveMembers returns the number of alive members
func (cm *ClusterManager) NumAliveMembers() int {
	alive := cm.GetAliveMembers()
	return len(alive)
}

// Shutdown shuts down the cluster manager
func (cm *ClusterManager) Shutdown() error {
	return cm.memberlist.Shutdown()
}

// EventDelegate implements memberlist.EventDelegate
type EventDelegate struct {
	manager *ClusterManager
}

// NotifyJoin is called when a node joins the cluster
func (e *EventDelegate) NotifyJoin(node *memberlist.Node) {
	member := nodeToMember(node)
	
	e.manager.mu.Lock()
	e.manager.members[member.Name] = member
	e.manager.mu.Unlock()

	// Send event
	select {
	case e.manager.eventCh <- memberlist.NodeEvent{
		Event: memberlist.NodeJoin,
		Node:  node,
	}:
	default:
		// Channel full, drop event
	}
}

// NotifyLeave is called when a node leaves the cluster
func (e *EventDelegate) NotifyLeave(node *memberlist.Node) {
	e.manager.mu.Lock()
	if member, exists := e.manager.members[node.Name]; exists {
		member.State = "dead"
	}
	e.manager.mu.Unlock()

	// Send event
	select {
	case e.manager.eventCh <- memberlist.NodeEvent{
		Event: memberlist.NodeLeave,
		Node:  node,
	}:
	default:
		// Channel full, drop event
	}
}

// NotifyUpdate is called when a node's metadata is updated
func (e *EventDelegate) NotifyUpdate(node *memberlist.Node) {
	member := nodeToMember(node)
	
	e.manager.mu.Lock()
	e.manager.members[member.Name] = member
	e.manager.mu.Unlock()

	// Send event
	select {
	case e.manager.eventCh <- memberlist.NodeEvent{
		Event: memberlist.NodeUpdate,
		Node:  node,
	}:
	default:
		// Channel full, drop event
	}
}

// NodeDelegate implementation

// NodeMeta returns metadata about this node
func (d *NodeDelegate) NodeMeta(limit int) []byte {
	meta := map[string]interface{}{
		"raft_addr": d.manager.config.RaftAddr,
		"version":   "1.0.0",
		"role":      "velocitydb",
	}

	data, _ := json.Marshal(meta)
	if len(data) > limit {
		return data[:limit]
	}
	return data
}

// NotifyMsg is called when a message is received
func (d *NodeDelegate) NotifyMsg([]byte) {}

// GetBroadcasts returns pending broadcasts
func (d *NodeDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

// LocalState returns local state for state synchronization
func (d *NodeDelegate) LocalState(join bool) []byte {
	state := map[string]interface{}{
		"node_id":    d.manager.localMember.ID,
		"raft_addr":  d.manager.localMember.RaftAddr,
		"timestamp":  time.Now().Unix(),
	}

	data, _ := json.Marshal(state)
	return data
}

// MergeRemoteState merges remote state during synchronization
func (d *NodeDelegate) MergeRemoteState(buf []byte, join bool) {
	var state map[string]interface{}
	if err := json.Unmarshal(buf, &state); err != nil {
		return
	}

	// Process remote state
	// This could be used to discover Raft addresses, etc.
}

// Helper functions

// nodeToMember converts a memberlist.Node to ClusterMember
func nodeToMember(node *memberlist.Node) *ClusterMember {
	member := &ClusterMember{
		ID:    node.Name,
		Name:  node.Name,
		Addr:  node.Addr.String(),
		Port:  node.Port,
		State: "alive",
		Meta:  make(map[string]string),
	}

	// Parse metadata
	if len(node.Meta) > 0 {
		var meta map[string]interface{}
		if err := json.Unmarshal(node.Meta, &meta); err == nil {
			if raftAddr, ok := meta["raft_addr"].(string); ok {
				member.RaftAddr = raftAddr
			}
			
			// Convert all metadata to strings
			for k, v := range meta {
				member.Meta[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	return member
}

// DiscoveryHelper helps with node discovery
type DiscoveryHelper struct {
	manager *ClusterManager
}

// NewDiscoveryHelper creates a new discovery helper
func NewDiscoveryHelper(manager *ClusterManager) *DiscoveryHelper {
	return &DiscoveryHelper{manager: manager}
}

// FindRaftPeers finds Raft peers in the cluster
func (d *DiscoveryHelper) FindRaftPeers() []string {
	members := d.manager.GetAliveMembers()
	var peers []string

	for _, member := range members {
		if member.RaftAddr != "" && member.Name != d.manager.localMember.Name {
			peers = append(peers, member.RaftAddr)
		}
	}

	return peers
}

// FindRaftLeader attempts to find the current Raft leader
func (d *DiscoveryHelper) FindRaftLeader() (string, error) {
	// This would need to query Raft nodes to find the leader
	// For now, return empty string
	return "", fmt.Errorf("leader discovery not implemented")
}

// WaitForMinimumMembers waits until the cluster has at least minMembers alive members
func (d *DiscoveryHelper) WaitForMinimumMembers(minMembers int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if d.manager.NumAliveMembers() >= minMembers {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	
	return fmt.Errorf("timeout waiting for minimum members (%d)", minMembers)
}