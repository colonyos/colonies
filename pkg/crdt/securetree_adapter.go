package crdt

type AdapterSecureNodeCRDT struct {
	nodeCrdt *NodeCRDT
}

func (n *AdapterSecureNodeCRDT) SetLiteral(value interface{}, prvKey string) error {
	clientID := ClientID(prvKey)
	return n.nodeCrdt.SetLiteral(value, clientID)
}

func (n *AdapterSecureNodeCRDT) GetLiteral() (interface{}, error) {
	return n.nodeCrdt.GetLiteral()
}

func (n *AdapterSecureNodeCRDT) CreateMapNode(prvKey string) (SecureNode, error) {
	clientID := ClientID(prvKey)
	node, err := n.nodeCrdt.CreateMapNode(clientID)
	if err != nil {
		return nil, err
	}
	return &AdapterSecureNodeCRDT{nodeCrdt: node}, nil
}

func (n *AdapterSecureNodeCRDT) SetKeyValue(key string, value interface{}, prvKey string) (NodeID, error) {
	clientID := ClientID(prvKey)
	return n.nodeCrdt.SetKeyValue(key, value, clientID)
}

func (n *AdapterSecureNodeCRDT) GetNodeForKey(key string) (SecureNode, bool, error) {
	internalNode, ok, err := n.nodeCrdt.GetNodeForKey(key)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &AdapterSecureNodeCRDT{nodeCrdt: internalNode}, ok, nil
}

func (n *AdapterSecureNodeCRDT) RemoveKeyValue(key string, prvKey string) error {
	clientID := ClientID(prvKey)
	return n.nodeCrdt.RemoveKeyValue(key, clientID)
}

type AdapterSecureTreeCRDT struct {
	treeCrdt *TreeCRDT
}

func NewSecureTree() SecureTree {
	return &AdapterSecureTreeCRDT{
		treeCrdt: newTreeCRDT(),
	}
}

func (c *AdapterSecureTreeCRDT) CreateAttachedNode(name string, isArray bool, parentID NodeID, prvKey string) SecureNode {
	clientID := ClientID(prvKey)
	node := c.treeCrdt.CreateAttachedNode(name, isArray, parentID, clientID)
	return &AdapterSecureNodeCRDT{nodeCrdt: node}
}

func (c *AdapterSecureTreeCRDT) CreateNode(name string, isArray bool, prvKey string) SecureNode {
	clientID := ClientID(prvKey)
	node := c.treeCrdt.CreateNode(name, isArray, clientID)
	return &AdapterSecureNodeCRDT{nodeCrdt: node}
}

func (c *AdapterSecureTreeCRDT) GetNode(id NodeID) (SecureNode, bool) {
	node, ok := c.treeCrdt.GetNode(id)
	if !ok {
		return nil, false
	}
	return &AdapterSecureNodeCRDT{nodeCrdt: node}, true
}

func (c *AdapterSecureTreeCRDT) GetSibling(parentNodeID NodeID, index int) (SecureNode, error) {
	node, err := c.treeCrdt.GetSibling(parentNodeID, index)
	if err != nil {
		return nil, err
	}
	return &AdapterSecureNodeCRDT{nodeCrdt: node}, nil
}

func (c *AdapterSecureTreeCRDT) GetValueByPath(path string) (interface{}, error) {
	return c.treeCrdt.GetValueByPath(path)
}

func (c *AdapterSecureTreeCRDT) GetNodeByPath(path string) (SecureNode, error) {
	node, err := c.treeCrdt.GetNodeByPath(path)
	if err != nil {
		return nil, err
	}
	return &AdapterSecureNodeCRDT{nodeCrdt: node}, nil
}

func (c *AdapterSecureTreeCRDT) GetStringValueByPath(path string) (string, error) {
	return c.treeCrdt.GetStringValueByPath(path)
}

func (c *AdapterSecureTreeCRDT) AddEdge(from, to NodeID, label string, prvKey string) error {
	clientID := ClientID(prvKey)
	return c.treeCrdt.AddEdge(from, to, label, clientID)
}

func (c *AdapterSecureTreeCRDT) RemoveEdge(from, to NodeID, prvKey string) error {
	clientID := ClientID(prvKey)
	return c.treeCrdt.RemoveEdge(from, to, clientID)
}

func (c *AdapterSecureTreeCRDT) AppendEdge(from, to NodeID, label string, prvKey string) error {
	clientID := ClientID(prvKey)
	return c.treeCrdt.AppendEdge(from, to, label, clientID)
}

func (c *AdapterSecureTreeCRDT) PrependEdge(from, to NodeID, label string, prvKey string) error {
	clientID := ClientID(prvKey)
	return c.treeCrdt.PrependEdge(from, to, label, clientID)
}

func (c *AdapterSecureTreeCRDT) InsertEdgeLeft(from, to NodeID, label string, sibling NodeID, prvKey string) {
	clientID := ClientID(prvKey)
	c.treeCrdt.InsertEdgeLeft(from, to, label, sibling, clientID)
}

func (c *AdapterSecureTreeCRDT) InsertEdgeRight(from, to NodeID, label string, sibling NodeID, prvKey string) error {
	clientID := ClientID(prvKey)
	return c.treeCrdt.InsertEdgeRight(from, to, label, sibling, clientID)
}

func (c *AdapterSecureTreeCRDT) Sync(c2 Tree, force bool) error {
	adapter, ok := c2.(*AdapterTreeCRDT)
	if !ok {
		panic("Merge: Tree must be of type *AdapterTreeCRDT")
	}
	return c.treeCrdt.Sync(adapter.treeCrdt, force)
}

func (c *AdapterSecureTreeCRDT) Merge(c2 Tree, force bool) error {
	adapter, ok := c2.(*AdapterTreeCRDT)
	if !ok {
		panic("Merge: Tree must be of type *AdapterTreeCRDT")
	}
	return c.treeCrdt.Merge(adapter.treeCrdt, force)
}

func (c *AdapterSecureTreeCRDT) ImportJSON(rawJSON []byte, parentID NodeID, edgeLabel string, idx int, isArray bool, prvKey string) (NodeID, error) {
	clientID := ClientID(prvKey)
	return c.treeCrdt.ImportJSON(rawJSON, parentID, edgeLabel, idx, isArray, clientID)
}

func (c *AdapterSecureTreeCRDT) ExportJSON() ([]byte, error) {
	return c.treeCrdt.ExportJSON()
}

func (c *AdapterSecureTreeCRDT) Load(data []byte) error {
	return c.treeCrdt.Load(data)
}

func (c *AdapterSecureTreeCRDT) Save() ([]byte, error) {
	return c.treeCrdt.Save()
}

func (c *AdapterSecureTreeCRDT) Tidy() {
	c.treeCrdt.Tidy()
}
