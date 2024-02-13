package p2p

import (
	"testing"
)

func TestCreateNode(t *testing.T) {
	expected := "Node{host123, [192.168.1.1, 10.0.0.1]}"
	node := CreateNode("host123", []string{"192.168.1.1", "10.0.0.1"})
	actual := node.String()

	if actual != expected {
		t.Errorf("CreateNode() = %v, want %v", actual, expected)
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
