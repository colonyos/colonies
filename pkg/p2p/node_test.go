package p2p

import (
	"reflect"
	"testing"
)

func TestCreateNode(t *testing.T) {
	hostID := "host123"
	addrs := []string{"192.168.1.1", "10.0.0.1"}

	expectedNode := &Node{
		HostID: hostID,
		Addr:   addrs,
	}

	// Actual node created by CreateNode
	node := CreateNode(hostID, addrs)

	if !reflect.DeepEqual(node, expectedNode) {
		t.Errorf("CreateNode() = %v, want %v", node, expectedNode)
	}
}

func TestNodeString(t *testing.T) {
	node := &Node{
		HostID: "host123",
		Addr:   []string{"192.168.1.1", "10.0.0.1"},
	}

	expectedStr := "Node{host123, [192.168.1.1, 10.0.0.1]}"

	actualStr := node.String()

	if actualStr != expectedStr {
		t.Errorf("String() = %v, want %v", actualStr, expectedStr)
	}
}

func TestNodeString2(t *testing.T) {
	node := &Node{
		HostID: "host123",
		Addr:   []string{"192.168.1.1"},
	}

	expectedStr := "Node{host123, [192.168.1.1]}"

	actualStr := node.String()

	if actualStr != expectedStr {
		t.Errorf("String() = %v, want %v", actualStr, expectedStr)
	}
}

func TestNodeString3(t *testing.T) {
	node := &Node{
		HostID: "host123",
		Addr:   []string{},
	}

	expectedStr := "Node{host123, []}"

	actualStr := node.String()

	if actualStr != expectedStr {
		t.Errorf("String() = %v, want %v", actualStr, expectedStr)
	}
}
