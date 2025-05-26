package crdt

type Node interface {
	// Literal operations
	SetLiteral(value interface{}, clientID ClientID) error
	GetLiteral() (interface{}, error)

	// Map operations
	CreateMapNode(clientID ClientID) (Node, error)
	SetKeyValue(key string, value interface{}, clientID ClientID) (NodeID, error)
	GetNodeForKey(key string) (Node, bool, error)
	RemoveKeyValue(key string, clientID ClientID) error
}

type Tree interface {
	// Node operations
	CreateAttachedNode(name string, isArray bool, parentID NodeID, clientID ClientID) Node
	CreateNode(name string, isArray bool, clientID ClientID) Node
	GetNode(id NodeID) (Node, bool)
	GetSibling(parentNodeID NodeID, index int) (Node, error)
	GetValueByPath(path string) (interface{}, error)
	GetNodeByPath(path string) (Node, error)
	GetStringValueByPath(path string) (string, error)

	// Edge operations
	AddEdge(from, to NodeID, label string, clientID ClientID) error
	RemoveEdge(from, to NodeID, clientID ClientID) error

	// List operations
	AppendEdge(from, to NodeID, label string, clientID ClientID) error
	PrependEdge(from, to NodeID, label string, clientID ClientID) error
	InsertEdgeLeft(from, to NodeID, label string, sibling NodeID, clientID ClientID)
	InsertEdgeRight(from, to NodeID, label string, sibling NodeID, clientID ClientID) error

	// Merge and operations
	Sync(c2 Tree, force bool) error
	Merge(c2 Tree, force bool) error

	// Serialization
	ImportJSON(rawJSON []byte, parentID NodeID, edgeLabel string, idx int, isArray bool, clientID ClientID) (NodeID, error)
	ExportJSON() ([]byte, error)
	Load(data []byte) error
	Save() ([]byte, error)

	// Utility functions
	Tidy()
}
