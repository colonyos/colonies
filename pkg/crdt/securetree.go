package crdt

type SecureNode interface {
	// General operations
	ID() NodeID

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
	// ABAC (Attribute Based Access Control)
	ABAC() *ABACPolicy

	// Node operations
	CreateAttachedNode(name string, nodeType NodeType, parentID NodeID, prvKey string) (SecureNode, error)
	CreateNode(name string, nodeType NodeType, prvKey string) (SecureNode, error)
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
	InsertEdgeLeft(from, to NodeID, label string, sibling NodeID, prvKey string) error
	InsertEdgeRight(from, to NodeID, label string, sibling NodeID, prvKey string) error

	// Merge operations
	Sync(c2 SecureTree, prvKey string) error
	Merge(c2 SecureTree, prvKey string) error

	// Serialization
	ImportJSON(rawJSON []byte, prvKey string) (NodeID, error)
	ImportJSONToMap(rawJSON []byte, parentID NodeID, key string, prvKey string) (NodeID, error)
	ImportJSONToArray(rawJSON []byte, parentID NodeID, prvKey string) (NodeID, error)
	ExportJSON() ([]byte, error)
	Load(data []byte) error
	Save() ([]byte, error)
	Clone() (SecureTree, error)

	// Utility functions
	Tidy()
}
