package crdt

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

type subscriber struct {
	path string
	ch   chan NodeEvent
}

type NodeEventType int

const (
	EventAdded NodeEventType = iota
	EventRemoved
	EventUpdated
	EventMarkedDeleted
)

type NodeEvent struct {
	NodeID NodeID
	Path   string
	Type   NodeEventType
}

func (c *TreeCRDT) Subscribe(path string, ch chan NodeEvent) {
	sub := subscriber{
		path: path,
		ch:   ch,
	}

	c.subscribers = append(c.subscribers, sub)

	log.WithFields(log.Fields{
		"Path": path,
	}).Debug("Subscribed to path")
}

func (c *TreeCRDT) computePath(nodeID NodeID) (string, error) {
	node, ok := c.Nodes[nodeID]
	if !ok {
		return "", fmt.Errorf("Node %s not found", nodeID)
	}

	var pathParts []string
	current := node
	for current != nil && !current.IsRoot {
		parentID := current.ParentID
		parent, exists := c.Nodes[parentID]
		if !exists {
			return "", fmt.Errorf("Parent %s not found for node %s", parentID, current.ID)
		}

		found := false
		for _, edge := range parent.Edges {
			if edge.To == current.ID {
				if parent.IsArray {
					// Array → index
					index := -1
					for i, e := range parent.Edges {
						if e.To == current.ID {
							index = i
							break
						}
					}
					if index >= 0 {
						pathParts = append([]string{fmt.Sprintf("%d", index)}, pathParts...)
					}
				} else {
					// Map → label
					if len(edge.Label) > 0 {
						pathParts = append([]string{edge.Label}, pathParts...)
					}
				}
				found = true
				break
			}
		}

		if !found {
			return "", fmt.Errorf("Could not compute path part for node %s", current.ID)
		}

		current = parent
	}

	path := "/" + strings.Join(pathParts, "/")
	return path, nil
}

func (c *TreeCRDT) notifySubscribers(nodeID NodeID, eventType NodeEventType) {
	nodePath, err := c.computePath(nodeID)
	if err != nil {
		log.WithFields(log.Fields{
			"NodeID": nodeID,
			"Error":  err,
		}).Error("Failed to compute path for notifySubscribers")
		return
	}

	evt := NodeEvent{
		NodeID: nodeID,
		Path:   nodePath,
		Type:   eventType,
	}

	for _, sub := range c.subscribers {
		if sub.path == nodePath || strings.HasPrefix(nodePath, sub.path) {
			select {
			case sub.ch <- evt:
				log.WithFields(log.Fields{
					"Path":   sub.path,
					"NodeID": nodeID,
					"Event":  eventType,
				}).Debug("Notified subscriber")
			default:
				log.WithFields(log.Fields{
					"Path":   sub.path,
					"NodeID": nodeID,
					"Event":  eventType,
				}).Warn("Subscriber channel full or blocked")
			}
		}
	}
}
