package crdt

import (
	"fmt"

	"github.com/colonyos/colonies/internal/crypto"
)

type AdapterSecureNodeCRDT struct {
	nodeCrdt *NodeCRDT
}

func performSecureAction(
	prvKey string,
	op string,
	action ABACAction,
	target NodeID,
	policy *ABACPolicy,
	actionFn func(ClientID) (*NodeCRDT, error),
) error {
	identity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return fmt.Errorf("failed to create identity: %w", err)
	}

	clientID := ClientID(identity.ID())

	if policy != nil && !policy.IsAllowed(clientID, action, target) {
		return fmt.Errorf("identity %s not allowed to perform %s on %s", clientID, action, target)
	}

	node, err := actionFn(clientID)
	if err != nil {
		return err
	}

	err = node.Sign(identity)
	if err != nil {
		return fmt.Errorf("failed to sign node: %w", err)
	}

	recoveredID, err := node.Verify()
	if err != nil {
		return fmt.Errorf("failed to verify node signature: %w", err)
	}
	if recoveredID != identity.ID() {
		return fmt.Errorf("signature verification failed: expected %s, got %s", identity.ID(), recoveredID)
	}

	return nil
}

func (n *AdapterSecureNodeCRDT) SetLiteral(value interface{}, prvKey string) error { // Tested
	op := buildOpString("Literal", value, n.nodeCrdt.ID)

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		if err := n.nodeCrdt.SetLiteral(value, clientID); err != nil {
			return nil, fmt.Errorf("failed to set literal: %w", err)
		}
		return n.nodeCrdt, nil
	}

	return performSecureAction(prvKey, op, ActionModify, n.nodeCrdt.ID, n.nodeCrdt.tree.ABACPolicy, secureAction)
}

func (n *AdapterSecureNodeCRDT) GetLiteral() (interface{}, error) {
	return n.nodeCrdt.GetLiteral()
}

func (n *AdapterSecureNodeCRDT) CreateMapNode(prvKey string) (SecureNode, error) { // Tested
	op := buildOpString("CreateMapNode")

	var newNode *NodeCRDT

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node, err := n.nodeCrdt.CreateMapNode(clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to create map node: %w", err)
		}
		newNode = node
		return newNode, nil
	}

	err := performSecureAction(prvKey, op, ActionAdd, n.nodeCrdt.ID, n.nodeCrdt.tree.ABACPolicy, secureAction)
	if err != nil {
		return nil, err
	}

	return &AdapterSecureNodeCRDT{nodeCrdt: newNode}, nil
}

func (n *AdapterSecureNodeCRDT) SetKeyValue(key string, value interface{}, prvKey string) (NodeID, error) { // Tested
	op := buildOpString("SetKeyValue", key, value)

	var newNodeID NodeID

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		id, err := n.nodeCrdt.SetKeyValue(key, value, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to set key-value: %w", err)
		}
		newNodeID = id
		return n.nodeCrdt, nil
	}

	err := performSecureAction(prvKey, op, ActionModify, n.nodeCrdt.ID, n.nodeCrdt.tree.ABACPolicy, secureAction)
	if err != nil {
		return "", err
	}

	return newNodeID, nil
}

func (n *AdapterSecureNodeCRDT) GetNodeForKey(key string) (SecureNode, bool, error) {
	internalNode, ok, err := n.nodeCrdt.GetNodeForKey(key)
	if err != nil || !ok {
		return nil, ok, err
	}
	return &AdapterSecureNodeCRDT{nodeCrdt: internalNode}, ok, nil
}

func (n *AdapterSecureNodeCRDT) RemoveKeyValue(key string, prvKey string) error { // Tested
	op := buildOpString("KeyValue", key)

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		if err := n.nodeCrdt.RemoveKeyValue(key, clientID); err != nil {
			return nil, fmt.Errorf("failed to remove key-value: %w", err)
		}
		return n.nodeCrdt, nil
	}

	return performSecureAction(
		prvKey,
		op,
		ActionRemove,
		n.nodeCrdt.ID,
		n.nodeCrdt.tree.ABACPolicy,
		secureAction,
	)
}

type AdapterSecureTreeCRDT struct {
	treeCrdt *TreeCRDT
}

func NewSecureTree(prvKey string) (SecureTree, error) {
	idendity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity from string: %w", err)
	}

	c := newTreeCRDT()
	c.OwnerID = idendity.ID()
	c.ABACPolicy = NewABACPolicy(c)
	c.ABACPolicy.Allow(ClientID(c.OwnerID), "*", "root", true) // Allow the owner to have full access to whole tree
	c.Secure = true

	return &AdapterSecureTreeCRDT{
		treeCrdt: c,
	}, nil
}

func (c *AdapterSecureTreeCRDT) CreateAttachedNode(name string, nodeType NodeType, parentID NodeID, prvKey string) (SecureNode, error) { // Tested
	var newNode *NodeCRDT

	op := buildOpString("Node", name, nodeType)

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node := c.treeCrdt.CreateAttachedNode(name, nodeType, parentID, clientID)
		newNode = node
		return newNode, nil
	}

	err := performSecureAction(
		prvKey,
		op,
		ActionAdd,
		parentID,
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
	if err != nil {
		return nil, err
	}

	return &AdapterSecureNodeCRDT{nodeCrdt: newNode}, nil
}

func (c *AdapterSecureTreeCRDT) CreateNode(name string, nodeType NodeType, prvKey string) (SecureNode, error) { // Tested
	var newNode *NodeCRDT

	op := buildOpString("Node", name, nodeType)

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node := c.treeCrdt.CreateNode(name, nodeType, clientID)
		newNode = node
		return newNode, nil
	}

	var nounce, signature string
	err := performSecureAction(
		prvKey,
		op,
		ActionAdd,
		c.treeCrdt.Root.ID, // Treat as adding under root
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
	if err != nil {
		return nil, err
	}

	newNode.Nounce = nounce
	newNode.Signature = signature

	return &AdapterSecureNodeCRDT{nodeCrdt: newNode}, nil
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

func (c *AdapterSecureTreeCRDT) AddEdge(from, to NodeID, label string, prvKey string) error { // Tested
	op := buildOpString("Edge", from, to, label, NodeID(""))

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		// Perform the actual edge addition
		node, ok := c.treeCrdt.GetNode(from)
		if !ok {
			return nil, fmt.Errorf("parent node %s not found", from)
		}

		err := c.treeCrdt.AddEdge(from, to, label, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to add edge from %s to %s: %w", from, to, err)
		}
		return node, nil
	}

	// Write to the parent's node.Nounce and node.Signature
	return performSecureAction(
		prvKey,
		op,
		ActionAdd,
		from, // ABAC checks and signing target is the parent node
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) RemoveEdge(from, to NodeID, prvKey string) error { // Tested
	op := buildOpString("RemoveEdge", from, to)

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node, ok := c.treeCrdt.GetNode(from)
		if !ok {
			return nil, fmt.Errorf("parent node %s not found", from)
		}
		err := c.treeCrdt.RemoveEdge(from, to, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to remove edge from %s to %s: %w", from, to, err)
		}
		return node, nil
	}

	return performSecureAction(
		prvKey,
		op,
		ActionRemove,
		from, // ABAC is enforced on the parent node
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) AppendEdge(from, to NodeID, label string, prvKey string) error { // Tested
	op := buildOpString("Edge", from, to, label)

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node, ok := c.treeCrdt.GetNode(from)
		if !ok {
			return nil, fmt.Errorf("parent node %s not found", from)
		}
		err := c.treeCrdt.AppendEdge(from, to, label, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to append edge from %s to %s: %w", from, to, err)
		}
		return node, nil
	}

	return performSecureAction(
		prvKey,
		op,
		ActionModify, // We treat appending a child as modifying the parent
		from,
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) PrependEdge(from, to NodeID, label string, prvKey string) error { // Tested
	op := buildOpString("Edge", from, to, label, NodeID(""))

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node, ok := c.treeCrdt.GetNode(from)
		if !ok {
			return nil, fmt.Errorf("parent node %s not found", from)
		}
		err := c.treeCrdt.PrependEdge(from, to, label, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to prepend edge from %s to %s: %w", from, to, err)
		}
		return node, nil
	}

	return performSecureAction(
		prvKey,
		op,
		ActionModify, // Modifying the parent node structure
		from,
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) InsertEdgeLeft(from, to NodeID, label string, sibling NodeID, prvKey string) error { // Tested
	op := buildOpString("Edge", from, to, label, sibling)

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node, ok := c.treeCrdt.Nodes[from]
		if !ok {
			return nil, fmt.Errorf("parent node %s not found", from)
		}
		err := c.treeCrdt.InsertEdgeLeft(from, to, label, sibling, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert edge left from %s to %s: %w", from, to, err)
		}
		return node, nil
	}

	return performSecureAction(
		prvKey,
		op,
		ActionModify,
		from,
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) InsertEdgeRight(from, to NodeID, label string, sibling NodeID, prvKey string) error {
	op := buildOpString("Edge", from, to, label, sibling)

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node, ok := c.treeCrdt.Nodes[from]
		if !ok {
			return nil, fmt.Errorf("parent node %s not found", from)
		}
		err := c.treeCrdt.InsertEdgeRight(from, to, label, sibling, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert edge right from %s to %s: %w", from, to, err)
		}
		return node, nil
	}

	return performSecureAction(
		prvKey,
		op,
		ActionModify,
		from,
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) Sync(c2 SecureTree, force bool) error { // TODO: test
	adapter, ok := c2.(*AdapterSecureTreeCRDT)
	if !ok {
		panic("Sync: Tree must be of type *AdapterTreeCRDT")
	}
	return c.treeCrdt.Sync(adapter.treeCrdt, force)
}

func (c *AdapterSecureTreeCRDT) Merge(c2 SecureTree, force bool) error { // TODO: test
	adapter, ok := c2.(*AdapterSecureTreeCRDT)
	if !ok {
		panic("Merge: Tree must be of type *AdapterTreeCRDT")
	}
	return c.treeCrdt.Merge(adapter.treeCrdt, force)
}

func (c *AdapterSecureTreeCRDT) ImportJSON(rawJSON []byte, prvKey string) (NodeID, error) { // Tested
	identity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", fmt.Errorf("failed to create identity from string: %w", err)
	}
	clientID := ClientID(identity.ID())

	if !c.treeCrdt.ABACPolicy.IsAllowed(clientID, ActionModify, c.treeCrdt.Root.ID) {
		return "", fmt.Errorf("identity %s is not allowed to import under root", clientID)
	}

	return c.treeCrdt.ImportJSON(rawJSON, clientID)
}

func (c *AdapterSecureTreeCRDT) ImportJSONToMap(rawJSON []byte, parentID NodeID, key string, prvKey string) (NodeID, error) { // Tested
	identity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", fmt.Errorf("failed to create identity from string: %w", err)
	}
	clientID := ClientID(identity.ID())

	if !c.treeCrdt.ABACPolicy.IsAllowed(clientID, ActionModify, parentID) {
		return "", fmt.Errorf("identity %s is not allowed to import under parent %s", clientID, parentID)
	}

	return c.treeCrdt.ImportJSONToMap(rawJSON, parentID, key, clientID)
}

func (c *AdapterSecureTreeCRDT) ImportJSONToArray(rawJSON []byte, parentID NodeID, prvKey string) (NodeID, error) {
	identity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", fmt.Errorf("failed to create identity from string: %w", err)
	}
	clientID := ClientID(identity.ID())

	if !c.treeCrdt.ABACPolicy.IsAllowed(clientID, ActionModify, parentID) {
		return "", fmt.Errorf("identity %s is not allowed to import under parent %s", clientID, parentID)
	}

	return c.treeCrdt.ImportJSONToArray(rawJSON, parentID, clientID)
}

func (c *AdapterSecureTreeCRDT) Clone() (SecureTree, error) {
	newTree, err := c.treeCrdt.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone tree: %w", err)
	}
	if newTree == nil {
		return nil, fmt.Errorf("failed to clone tree")
	}
	return &AdapterSecureTreeCRDT{treeCrdt: newTree}, nil
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
