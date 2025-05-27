package crdt

type SecureNode interface {
	// Literal operations
	SetLiteral(value interface{}, prvKey string) error
	GetLiteral() (interface{}, error)

	// Map operations
	CreateMapNode(prvKey string) (SecureNode, error)
	SetKeyValue(key string, value interface{}, prvKey string) (NodeID, error)
	GetNodeForKey(key string) (SecureNode, bool, error)
	RemoveKeyValue(key string, prvKey string) error
}

type SecureTree interface {
	// Node operations
	CreateAttachedNode(name string, isArray bool, parentID NodeID, prvKey string) SecureNode
	CreateNode(name string, isArray bool, prvKey string) SecureNode
	GetNode(id NodeID) (SecureNode, bool)
	GetSibling(parentNodeID NodeID, index int) (SecureNode, error)
	GetValueByPath(path string) (interface{}, error)
	GetNodeByPath(path string) (SecureNode, error)
	GetStringValueByPath(path string) (string, error)

	// Edge operations
	AddEdge(from, to NodeID, label string, prvKey string) error
	RemoveEdge(from, to NodeID, prvKey string) error

	// List operations
	AppendEdge(from, to NodeID, label string, prvKey string) error
	PrependEdge(from, to NodeID, label string, prvKey string) error
	InsertEdgeLeft(from, to NodeID, label string, sibling NodeID, prvKey string)
	InsertEdgeRight(from, to NodeID, label string, sibling NodeID, prvKey string) error

	// Merge and operations
	Sync(c2 Tree, force bool) error
	Merge(c2 Tree, force bool) error

	// Serialization
	ImportJSON(rawJSON []byte, parentID NodeID, edgeLabel string, idx int, isArray bool, prvKey string) (NodeID, error)
	ExportJSON() ([]byte, error)
	Load(data []byte) error
	Save() ([]byte, error)

	// Utility functions
	Tidy()
}
