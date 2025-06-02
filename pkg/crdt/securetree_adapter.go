package crdt

import (
	"fmt"

	"github.com/colonyos/colonies/internal/crypto"
	log "github.com/sirupsen/logrus"
)

type AdapterSecureNodeCRDT struct {
	nodeCrdt *NodeCRDT
}

func performSecureAction(
	accessControl bool,
	prvKey string,
	action ABACAction,
	target NodeID,
	abac *ABACPolicy,
	actionFn func(ClientID) (*NodeCRDT, error),
) error {
	identity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return fmt.Errorf("failed to create identity: %w", err)
	}

	id := identity.ID()

	if accessControl {
		if abac != nil && !abac.IsAllowed(id, action, target) {
			log.WithFields(log.Fields{
				"ID":     id,
				"Action": action,
				"Target": target,
			}).Error("Not allowed to perform action on target")
			return fmt.Errorf("identity %s not allowed to perform %s on %s", id, action, target)
		}
	}

	node, err := actionFn(ClientID(id))
	if err != nil {
		return err
	}

	err = node.Sign(identity)
	if err != nil {
		return fmt.Errorf("failed to sign node: %w", err)
	}

	if node.ParentID != "" {
		parentNode, ok := node.tree.GetNode(node.ParentID)
		if !ok {
			return fmt.Errorf("parent node %s not found for node %s", node.ParentID, node.ID)
		}

		// Sign the parent node with the same identity
		if err := parentNode.Sign(identity); err != nil {
			return fmt.Errorf("failed to sign parent node: %w", err)
		}
	}

	return nil
}

func (n *AdapterSecureNodeCRDT) ID() NodeID {
	return n.nodeCrdt.ID
}

func (n *AdapterSecureNodeCRDT) SetLiteral(value interface{}, prvKey string) error { // Tested
	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		if err := n.nodeCrdt.SetLiteral(value, clientID); err != nil {
			return nil, fmt.Errorf("failed to set literal: %w", err)
		}
		return n.nodeCrdt, nil
	}

	accessControl := true
	if n.nodeCrdt.ParentID == "" {
		accessControl = false // If the node is not attached to a tree, we skip ABAC checks
	}
	return performSecureAction(
		accessControl,
		prvKey,
		ActionModify,
		n.nodeCrdt.ID,
		n.nodeCrdt.tree.ABACPolicy,
		secureAction)
}

func (n *AdapterSecureNodeCRDT) GetLiteral() (interface{}, error) {
	return n.nodeCrdt.GetLiteral()
}

func (n *AdapterSecureNodeCRDT) CreateMapNode(prvKey string) (SecureNode, error) { // Tested
	var newNode *NodeCRDT

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node, err := n.nodeCrdt.CreateMapNode(clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to create map node: %w", err)
		}
		newNode = node
		return newNode, nil
	}

	err := performSecureAction(
		true,
		prvKey,
		ActionModify,
		n.nodeCrdt.ID,
		n.nodeCrdt.tree.ABACPolicy,
		secureAction)
	if err != nil {
		return nil, err
	}

	return &AdapterSecureNodeCRDT{nodeCrdt: newNode}, nil
}

func (n *AdapterSecureNodeCRDT) SetKeyValue(key string, value interface{}, prvKey string) (NodeID, error) { // Tested
	var newNodeID NodeID

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		id, err := n.nodeCrdt.SetKeyValue(key, value, clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to set key-value: %w", err)
		}
		newNodeID = id
		newNode, ok := n.nodeCrdt.tree.GetNode(newNodeID)
		if !ok {
			return nil, fmt.Errorf("new node %s not found in tree after setting key-value", newNodeID)
		}
		return newNode, nil
	}

	err := performSecureAction(
		true,
		prvKey,
		ActionModify,
		n.nodeCrdt.ID,
		n.nodeCrdt.tree.ABACPolicy,
		secureAction)
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
	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		if err := n.nodeCrdt.RemoveKeyValue(key, clientID); err != nil {
			return nil, fmt.Errorf("failed to remove key-value: %w", err)
		}
		return n.nodeCrdt, nil
	}

	return performSecureAction(
		true,
		prvKey,
		ActionModify,
		n.nodeCrdt.ID,
		n.nodeCrdt.tree.ABACPolicy,
		secureAction,
	)
}

type AdapterSecureTreeCRDT struct {
	treeCrdt *TreeCRDT
}

func NewSecureTree(prvKey string) (SecureTree, error) {
	identity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity from string: %w", err)
	}

	c := newTreeCRDT()
	ownerID := identity.ID()
	c.ABACPolicy = NewABACPolicy(c, ownerID, identity)
	c.ABACPolicy.Allow(ownerID, "*", "root", true) // Allow the owner to have full access to whole tree
	c.Secure = true

	return &AdapterSecureTreeCRDT{
		treeCrdt: c,
	}, nil
}

func (c *AdapterSecureTreeCRDT) ABAC() *ABACPolicy {
	return c.treeCrdt.ABACPolicy
}

func (c *AdapterSecureTreeCRDT) CreateAttachedNode(name string, nodeType NodeType, parentID NodeID, prvKey string) (SecureNode, error) { // Tested
	var newNode *NodeCRDT

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node := c.treeCrdt.CreateAttachedNode(name, nodeType, parentID, clientID)
		newNode = node
		return newNode, nil
	}

	err := performSecureAction(
		true,
		prvKey,
		ActionModify,
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

	secureAction := func(clientID ClientID) (*NodeCRDT, error) {
		node := c.treeCrdt.CreateNode(name, nodeType, clientID)
		newNode = node
		return newNode, nil
	}

	var nounce, signature string
	err := performSecureAction(
		false, // Check ABAC policy since this node is not attached to the tree yet
		prvKey,
		ActionModify,
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
		true,
		prvKey,
		ActionModify,
		from, // ABAC checks and signing target is the parent node
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) RemoveEdge(from, to NodeID, prvKey string) error { // Tested
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
		true,
		prvKey,
		ActionModify,
		from, // ABAC is enforced on the parent node
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) AppendEdge(from, to NodeID, label string, prvKey string) error { // Tested
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
		true,
		prvKey,
		ActionModify, // We treat appending a child as modifying the parent
		from,
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) PrependEdge(from, to NodeID, label string, prvKey string) error { // Tested
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
		true,
		prvKey,
		ActionModify, // Modifying the parent node structure
		from,
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) InsertEdgeLeft(from, to NodeID, label string, sibling NodeID, prvKey string) error { // Tested
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
		true,
		prvKey,
		ActionModify,
		from,
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) InsertEdgeRight(from, to NodeID, label string, sibling NodeID, prvKey string) error {
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
		true,
		prvKey,
		ActionModify,
		from,
		c.treeCrdt.ABACPolicy,
		secureAction,
	)
}

func (c *AdapterSecureTreeCRDT) Merge(c2 SecureTree, prvKey string) error { // TODO: test
	adapter, ok := c2.(*AdapterSecureTreeCRDT)
	if !ok {
		panic("Merge: Tree must be of type *AdapterTreeCRDT")
	}
	return c.treeCrdt.SecureMerge(adapter.treeCrdt, prvKey)
}

func (c *AdapterSecureTreeCRDT) ImportJSON(rawJSON []byte, prvKey string) (NodeID, error) { // Tested
	identity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", fmt.Errorf("failed to create identity from string: %w", err)
	}

	id := identity.ID()

	if !c.treeCrdt.ABACPolicy.IsAllowed(id, ActionModify, c.treeCrdt.Root.ID) {
		return "", fmt.Errorf("identity %s is not allowed to import under root", id)
	}

	return c.treeCrdt.SecureImportJSON(rawJSON, identity)
}

func (c *AdapterSecureTreeCRDT) ImportJSONToMap(rawJSON []byte, parentID NodeID, key string, prvKey string) (NodeID, error) { // Tested
	identity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", fmt.Errorf("failed to create identity from string: %w", err)
	}

	id := identity.ID()

	if !c.treeCrdt.ABACPolicy.IsAllowed(id, ActionModify, parentID) {
		return "", fmt.Errorf("identity %s is not allowed to import under parent %s", id, parentID)
	}

	return c.treeCrdt.SecureImportJSONToMap(rawJSON, parentID, key, identity)
}

func (c *AdapterSecureTreeCRDT) ImportJSONToArray(rawJSON []byte, parentID NodeID, prvKey string) (NodeID, error) {
	identity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", fmt.Errorf("failed to create identity from string: %w", err)
	}

	id := identity.ID()

	if !c.treeCrdt.ABACPolicy.IsAllowed(id, ActionModify, parentID) {
		return "", fmt.Errorf("identity %s is not allowed to import under parent %s", id, parentID)
	}

	return c.treeCrdt.SecureImportJSONToArray(rawJSON, parentID, identity)
}

func (c *AdapterSecureTreeCRDT) Clone() (SecureTree, error) {
	newTree, err := c.treeCrdt.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone tree: %w", err)
	}
	if newTree == nil {
		return nil, fmt.Errorf("failed to clone tree")
	}
	newTree.ABACPolicy, err = newTree.ABACPolicy.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone ABAC policy: %w", err)
	}
	newTree.ABACPolicy.tree = newTree
	newTree.ABACPolicy.identity = c.treeCrdt.ABACPolicy.identity
	return &AdapterSecureTreeCRDT{treeCrdt: newTree}, nil
}

func (c *AdapterSecureTreeCRDT) ExportJSON() ([]byte, error) {
	return c.treeCrdt.ExportJSON()
}

func (c *AdapterSecureTreeCRDT) Load(data []byte) error {
	identity := c.treeCrdt.ABACPolicy.identity
	err := c.treeCrdt.Load(data)
	if err != nil {
		return fmt.Errorf("failed to load tree data: %w", err)
	}
	c.treeCrdt.ABACPolicy.tree = c.treeCrdt
	c.treeCrdt.ABACPolicy.identity = identity
	recoveredID, err := c.treeCrdt.ABACPolicy.Verify()
	if err != nil {
		log.WithFields(log.Fields{
			"Identity":    identity.ID(),
			"Action":      "Load",
			"Owner":       c.treeCrdt.ABACPolicy.OwnerID,
			"RecoveredID": recoveredID,
			"Error":       err,
		}).Error("Failed to verify ABAC policy after loading, recovered ID does not match owner ID")

		return fmt.Errorf("failed to verify ABAC policy after loading: %w", err)
	}

	return nil
}

func (c *AdapterSecureTreeCRDT) Save() ([]byte, error) {
	return c.treeCrdt.Save()
}

func (c *AdapterSecureTreeCRDT) Subscribe(path string, ch chan NodeEvent) {
	c.treeCrdt.Subscribe(path, ch)
}

func (c *AdapterSecureTreeCRDT) Tidy() {
	c.treeCrdt.Tidy()
}
