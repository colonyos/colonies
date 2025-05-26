package crdt

type AdapterNodeCRDT struct {
	nodeCrdt *NodeCRDT
}

func (n *AdapterNodeCRDT) SetLiteral(value interface{}, clientID ClientID) error {
	return n.nodeCrdt.SetLiteral(value, clientID)
}

func (n *AdapterNodeCRDT) GetLiteral() (interface{}, error) {
	return n.nodeCrdt.GetLiteral()
}

func (n *AdapterNodeCRDT) CreateMapNode(clientID ClientID) (Node, error) {
	node, err := n.nodeCrdt.CreateMapNode(clientID)
	if err != nil {
		return nil, err
	}
	return &AdapterNodeCRDT{nodeCrdt: node}, nil
}

func (n *AdapterNodeCRDT) SetKeyValue(key string, value interface{}, clientID ClientID) (NodeID, error) {
	return n.nodeCrdt.SetKeyValue(key, value, clientID)
}

func (n *AdapterNodeCRDT) GetNodeForKey(key string) (Node, bool, error) {
	internalNode, ok, err := n.nodeCrdt.GetNodeForKey(key)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &AdapterNodeCRDT{nodeCrdt: internalNode}, ok, nil
}

func (n *AdapterNodeCRDT) RemoveKeyValue(key string, clientID ClientID) error {
	return n.nodeCrdt.RemoveKeyValue(key, clientID)
}

type AdapterTreeCRDT struct {
	treeCrdt *TreeCRDT
}

func NewTree() Tree {
	return &AdapterTreeCRDT{
		treeCrdt: NewTreeCRDT(),
	}
}

func (c *AdapterTreeCRDT) CreateAttachedNode(name string, isArray bool, parentID NodeID, clientID ClientID) Node {
	node := c.treeCrdt.CreateAttachedNode(name, isArray, parentID, clientID)
	return &AdapterNodeCRDT{nodeCrdt: node}
}

func (c *AdapterTreeCRDT) CreateNode(name string, isArray bool, clientID ClientID) Node {
	node := c.treeCrdt.CreateNode(name, isArray, clientID)
	return &AdapterNodeCRDT{nodeCrdt: node}
}

func (c *AdapterTreeCRDT) GetNode(id NodeID) (Node, bool) {
	node, ok := c.treeCrdt.GetNode(id)
	if !ok {
		return nil, false
	}
	return &AdapterNodeCRDT{nodeCrdt: node}, true
}

func (c *AdapterTreeCRDT) GetSibling(parentNodeID NodeID, index int) (Node, error) {
	node, err := c.treeCrdt.GetSibling(parentNodeID, index)
	if err != nil {
		return nil, err
	}
	return &AdapterNodeCRDT{nodeCrdt: node}, nil
}

func (c *AdapterTreeCRDT) GetValueByPath(path string) (interface{}, error) {
	return c.treeCrdt.GetValueByPath(path)
}

func (c *AdapterTreeCRDT) GetNodeByPath(path string) (Node, error) {
	node, err := c.treeCrdt.GetNodeByPath(path)
	if err != nil {
		return nil, err
	}
	return &AdapterNodeCRDT{nodeCrdt: node}, nil
}

func (c *AdapterTreeCRDT) GetStringValueByPath(path string) (string, error) {
	return c.treeCrdt.GetStringValueByPath(path)
}

func (c *AdapterTreeCRDT) AddEdge(from, to NodeID, label string, clientID ClientID) error {
	return c.treeCrdt.AddEdge(from, to, label, clientID)
}

func (c *AdapterTreeCRDT) RemoveEdge(from, to NodeID, clientID ClientID) error {
	return c.treeCrdt.RemoveEdge(from, to, clientID)
}

func (c *AdapterTreeCRDT) AppendEdge(from, to NodeID, label string, clientID ClientID) error {
	return c.treeCrdt.AppendEdge(from, to, label, clientID)
}

func (c *AdapterTreeCRDT) PrependEdge(from, to NodeID, label string, clientID ClientID) error {
	return c.treeCrdt.PrependEdge(from, to, label, clientID)
}

func (c *AdapterTreeCRDT) InsertEdgeLeft(from, to NodeID, label string, sibling NodeID, clientID ClientID) {
	c.treeCrdt.InsertEdgeLeft(from, to, label, sibling, clientID)
}

func (c *AdapterTreeCRDT) InsertEdgeRight(from, to NodeID, label string, sibling NodeID, clientID ClientID) error {
	return c.treeCrdt.InsertEdgeRight(from, to, label, sibling, clientID)
}

func (c *AdapterTreeCRDT) Sync(c2 Tree, force bool) error {
	adapter, ok := c2.(*AdapterTreeCRDT)
	if !ok {
		panic("Merge: Tree must be of type *AdapterTreeCRDT")
	}
	return c.treeCrdt.Sync(adapter.treeCrdt, force)
}

func (c *AdapterTreeCRDT) Merge(c2 Tree, force bool) error {
	adapter, ok := c2.(*AdapterTreeCRDT)
	if !ok {
		panic("Merge: Tree must be of type *AdapterTreeCRDT")
	}
	return c.treeCrdt.Merge(adapter.treeCrdt, force)
}

func (c *AdapterTreeCRDT) ImportJSON(rawJSON []byte, parentID NodeID, edgeLabel string, idx int, isArray bool, clientID ClientID) (NodeID, error) {
	return c.treeCrdt.ImportJSON(rawJSON, parentID, edgeLabel, idx, isArray, clientID)
}

func (c *AdapterTreeCRDT) ExportJSON() ([]byte, error) {
	return c.treeCrdt.ExportJSON()
}

func (c *AdapterTreeCRDT) Load(data []byte) error {
	return c.treeCrdt.Load(data)
}

func (c *AdapterTreeCRDT) Save() ([]byte, error) {
	return c.treeCrdt.Save()
}

func (c *AdapterTreeCRDT) Tidy() {
	c.treeCrdt.Tidy()
}
