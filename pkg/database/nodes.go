package database

import "github.com/colonyos/colonies/pkg/core"

type NodeDatabase interface {
	AddNode(node *core.Node) error
	GetNodeByID(nodeID string) (*core.Node, error)
	GetNodeByName(colonyName string, nodeName string) (*core.Node, error)
	GetNodes(colonyName string) ([]*core.Node, error)
	GetNodesByLocation(colonyName string, location string) ([]*core.Node, error)
	UpdateNode(node *core.Node) error
	RemoveNodeByID(nodeID string) error
	RemoveNodeByName(colonyName string, nodeName string) error
	RemoveNodesByColonyName(colonyName string) error
	CountNodes(colonyName string) (int, error)
}
