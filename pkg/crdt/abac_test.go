package crdt

import "testing"

type mockTree struct{}

func (m *mockTree) IsDescendant(root NodeID, target NodeID) bool {
	return root == "parent" && target == "child"
}

func TestABACPolicyWithWildcardsAndRecursive(t *testing.T) {
	tree := &mockTree{}
	policy := NewABACPolicy(tree)

	clientA := ClientID("alice")
	clientB := ClientID("bob")
	nodeX := NodeID("node-x")
	nodeY := NodeID("node-y")
	parent := NodeID("parent")
	child := NodeID("child")

	// Setup rules
	policy.Allow("*", ActionAdd, "*", false)          // Anyone can add anywhere
	policy.Allow(clientA, ActionModify, "*", false)   // Alice can modify anything
	policy.Allow("*", ActionRemove, nodeX, false)     // Anyone can remove node-x
	policy.Allow(clientB, "*", nodeY, false)          // Bob can do anything on node-y
	policy.Allow(clientA, ActionModify, parent, true) // Alice can modify parent and children

	tests := []struct {
		name     string
		clientID ClientID
		action   ABACAction
		nodeID   NodeID
		expected bool
	}{
		{"Wildcard add (alice on node-x)", clientA, ActionAdd, nodeX, true},
		{"Wildcard add (bob on node-y)", clientB, ActionAdd, nodeY, true},
		{"Wildcard modify (alice on node-x)", clientA, ActionModify, nodeX, true},
		{"Wildcard modify (bob on node-x)", clientB, ActionModify, nodeX, false},
		{"Wildcard remove (charlie on node-x)", "charlie", ActionRemove, nodeX, true},
		{"Wildcard remove (charlie on node-y)", "charlie", ActionRemove, nodeY, false},
		{"Wildcard action (bob on node-y, remove)", clientB, ActionRemove, nodeY, true},
		{"Wildcard action (bob on node-y, modify)", clientB, ActionModify, nodeY, true},
		{"No rule (charlie on node-y, modify)", "charlie", ActionModify, nodeY, false},
		{"Recursive rule (alice on child of parent)", clientA, ActionModify, child, true},
		{"Recursive rule (bob on child of parent)", clientB, ActionModify, child, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := policy.IsAllowed(test.clientID, test.action, test.nodeID)
			if result != test.expected {
				t.Errorf("IsAllowed(%s, %s, %s) = %v; want %v", test.clientID, test.action, test.nodeID, result, test.expected)
			}
		})
	}
}

func TestABACPolicyUpdateAndRemove(t *testing.T) {
	tree := &mockTree{}
	policy := NewABACPolicy(tree)
	client := ClientID("carol")
	node := NodeID("node-test")

	// Initially deny
	if policy.IsAllowed(client, ActionModify, node) {
		t.Errorf("Expected not allowed before rule is added")
	}

	// Add and verify
	policy.Allow(client, ActionModify, node, false)
	if !policy.IsAllowed(client, ActionModify, node) {
		t.Errorf("Expected allowed after rule is added")
	}

	// Update to recursive
	policy.UpdateRule(client, ActionModify, node, true)
	rule := policy.Rules[client][ActionModify][node]
	if !rule.Recursive {
		t.Errorf("Expected rule to be recursive after update")
	}

	// Remove and verify
	policy.RemoveRule(client, ActionModify, node)
	if policy.IsAllowed(client, ActionModify, node) {
		t.Errorf("Expected not allowed after rule is removed")
	}
}
