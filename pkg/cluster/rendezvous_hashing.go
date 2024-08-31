package cluster

import (
	"crypto/sha256"
	"encoding/binary"
)

type RendezvousHash struct {
	nodes []string
}

func NewRendezvousHash(nodes []string) *RendezvousHash {
	return &RendezvousHash{
		nodes: nodes,
	}
}

func hash(key, node string) uint64 {
	h := sha256.New()
	h.Write([]byte(key + node))
	hashBytes := h.Sum(nil)
	return binary.BigEndian.Uint64(hashBytes[:8])
}

func (rh *RendezvousHash) GetNode(key string) string {
	var maxWeight uint64
	var selectedNode string
	for _, node := range rh.nodes {
		weight := hash(key, node)
		if weight > maxWeight || selectedNode == "" {
			maxWeight = weight
			selectedNode = node
		}
	}
	return selectedNode
}
