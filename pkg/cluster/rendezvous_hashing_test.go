package cluster

import (
	"testing"
)

func TestRendezVousHahing(t *testing.T) {
	nodes := []string{"Node1", "Node2", "Node3", "Node4"}
	rh := NewRendezvousHash(nodes)

	keys := []string{"Key1", "Key2", "Key3", "Key4", "Key5", "Key6", "Key7", "Key8", "Key9", "Key10"}

	for _, key := range keys {
		node := rh.GetNode(key)

		if key == "Key1" {
			if node != "Node1" {
				t.Errorf("Expected Node1, got %s", node)
			}
		} else if key == "Key2" {
			if node != "Node2" {
				t.Errorf("Expected Node2, got %s", node)
			}
		} else if key == "Key3" {
			if node != "Node1" {
				t.Errorf("Expected Node1, got %s", node)
			}
		} else if key == "Key4" {
			if node != "Node4" {
				t.Errorf("Expected Node4, got %s", node)
			}
		} else if key == "Key5" {
			if node != "Node3" {
				t.Errorf("Expected Node3, got %s", node)
			}
		} else if key == "Key6" {
			if node != "Node2" {
				t.Errorf("Expected Node2, got %s", node)
			}
		} else if key == "Key7" {
			if node != "Node3" {
				t.Errorf("Expected Node3, got %s", node)
			}
		} else if key == "Key8" {
			if node != "Node4" {
				t.Errorf("Expected Node4, got %s", node)
			}
		} else if key == "Key9" {
			if node != "Node4" {
				t.Errorf("Expected Node4, got %s", node)
			}
		} else if key == "Key10" {
			if node != "Node1" {
				t.Errorf("Expected Node1, got %s", node)
			}
		}
	}
}
