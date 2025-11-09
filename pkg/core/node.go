package core

import (
	"encoding/json"
	"time"

	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/google/uuid"
)

// Node represents a physical or virtual host in the colony
type Node struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"` // Hostname
	ColonyName   string            `json:"colonyname"`
	Location     string            `json:"location"`     // User-provided location/zone
	Platform     string            `json:"platform"`     // "linux", "darwin", "windows"
	Architecture string            `json:"architecture"` // "amd64", "arm64", "riscv64"
	CPU          int               `json:"cpu"`          // Number of CPU cores
	Memory       int64             `json:"memory"`       // Memory in MB
	GPU          int               `json:"gpu"`          // Number of GPUs
	Capabilities []string          `json:"capabilities"` // ["docker", "gpu", "privileged"]
	Labels       map[string]string `json:"labels"`
	State        string            `json:"state"` // "ready", "not-ready", "offline"
	LastSeen     time.Time         `json:"lastseen"`
	Created      time.Time         `json:"created"`
}

// NodeMetadata contains node information attached to executor registration
type NodeMetadata struct {
	Hostname     string            `json:"hostname"`
	Location     string            `json:"location"`
	Platform     string            `json:"platform"`
	Architecture string            `json:"architecture"`
	CPU          int               `json:"cpu"`
	Memory       int64             `json:"memory"`
	GPU          int               `json:"gpu"`
	Capabilities []string          `json:"capabilities"`
	Labels       map[string]string `json:"labels"`
}

// CreateNode creates a new node
func CreateNode(name, colonyName, location string) *Node {
	nodeUUID := uuid.New()
	nodeCrypto := crypto.CreateCrypto()
	id := nodeCrypto.GenerateHash(nodeUUID.String())

	node := &Node{
		ID:           id,
		Name:         name,
		ColonyName:   colonyName,
		Location:     location,
		Labels:       make(map[string]string),
		Capabilities: make([]string, 0),
		State:        "ready",
		Created:      time.Now(),
		LastSeen:     time.Now(),
	}

	return node
}

// Equals compares two nodes
func (node *Node) Equals(node2 *Node) bool {
	if node2 == nil {
		return false
	}

	if node.ID == node2.ID &&
		node.Name == node2.Name &&
		node.ColonyName == node2.ColonyName &&
		node.Location == node2.Location &&
		node.Platform == node2.Platform &&
		node.Architecture == node2.Architecture {
		return true
	}

	return false
}

// ToJSON converts the node to JSON
func (node *Node) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(node, "", "    ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// ConvertJSONToNode converts JSON to a node
func ConvertJSONToNode(jsonString string) (*Node, error) {
	var node *Node
	err := json.Unmarshal([]byte(jsonString), &node)
	if err != nil {
		return nil, err
	}

	return node, nil
}

// ConvertJSONToNodes converts JSON to nodes
func ConvertJSONToNodes(jsonString string) ([]*Node, error) {
	var nodes []*Node
	err := json.Unmarshal([]byte(jsonString), &nodes)
	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// ConvertNodesToJSON converts nodes to JSON
func ConvertNodesToJSON(nodes []*Node) (string, error) {
	jsonBytes, err := json.Marshal(nodes)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// TouchLastSeen updates the last seen time for the node
func (node *Node) TouchLastSeen() {
	node.LastSeen = time.Now()
}

// UpdateFromMetadata updates node fields from NodeMetadata
func (node *Node) UpdateFromMetadata(metadata *NodeMetadata) {
	if metadata == nil {
		return
	}

	node.Platform = metadata.Platform
	node.Architecture = metadata.Architecture
	node.CPU = metadata.CPU
	node.Memory = metadata.Memory
	node.GPU = metadata.GPU
	node.Capabilities = metadata.Capabilities

	if metadata.Location != "" {
		node.Location = metadata.Location
	}

	// Merge labels
	if metadata.Labels != nil {
		for k, v := range metadata.Labels {
			node.Labels[k] = v
		}
	}

	node.State = "ready"
	node.LastSeen = time.Now()
}
