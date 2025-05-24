package crdt

type Tree interface {
	// Node and Edge operations
	CreateAttachedNode(name string, isArray bool, parentID NodeID, clientID ClientID) *Node
	CreateNode(name string, isArray bool, clientID ClientID) *Node
	GetNode(id NodeID) (*Node, bool)
	AddEdge(from, to NodeID, label string, clientID ClientID) error
	RemoveEdge(from, to NodeID, clientID ClientID) error

	// Literal operations
	SetLiteral(value interface{}, clientID ClientID, version int) error

	// Map operations
	GetValue(key string) (interface{}, bool)
	GetValueByPath(path string) (interface{}, bool, error)
	GetStringValueByPath(path string) (string, bool, error)
	SetField(key string, value interface{}, clientID ClientID, version int)
	RemoveField(key string, clientID ClientID, version int)

	// List operations
	AppendEdge(from, to NodeID, label string, clientID ClientID) error
	PrependEdge(from, to NodeID, label string, clientID ClientID) error
	InsertEdgeLeft(from, to NodeID, label string, sibling NodeID, clientID ClientID)
	InsertEdgeRight(from, to NodeID, label string, sibling NodeID, clientID ClientID) error
	GetSibling(parentNodeID NodeID, index int) (*Node, error)

	// Merge operations
	Merge(c2 *TreeCRDT)

	// Serialization
	ImportJSON(rawJSON []byte, parentID NodeID, edgeLabel string, idx int, isArray bool, clientID ClientID) (NodeID, error)
	ExportJSON() ([]byte, error)
	Load(data []byte) error
	Save() ([]byte, error)

	// Utility functions
	Tidy()
	Print()
}
