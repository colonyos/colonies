package crdt

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type ABACAction string

const (
	ActionAdd    ABACAction = "add"
	ActionModify ABACAction = "modify"
	ActionRemove ABACAction = "remove"
)

type ABACRule struct {
	Recursive bool `json:"recursive"`
}

type TreeChecker interface {
	isDescendant(root NodeID, target NodeID) bool
}

type ABACPolicy struct {
	Rules map[ClientID]map[ABACAction]map[NodeID]ABACRule `json:"rules"`
	tree  TreeChecker                                     `json:"-"`
}

func NewABACPolicy(tree TreeChecker) *ABACPolicy {
	return &ABACPolicy{
		Rules: make(map[ClientID]map[ABACAction]map[NodeID]ABACRule),
		tree:  tree,
	}
}

func (p *ABACPolicy) SetTree(tree TreeChecker) {
	p.tree = tree
}

func (p *ABACPolicy) Allow(clientID ClientID, action ABACAction, nodeID NodeID, recursive bool) {
	if _, ok := p.Rules[clientID]; !ok {
		p.Rules[clientID] = make(map[ABACAction]map[NodeID]ABACRule)
	}
	if _, ok := p.Rules[clientID][action]; !ok {
		p.Rules[clientID][action] = make(map[NodeID]ABACRule)
	}
	p.Rules[clientID][action][nodeID] = ABACRule{Recursive: recursive}
}

func (p *ABACPolicy) UpdateRule(clientID ClientID, action ABACAction, nodeID NodeID, recursive bool) {
	p.Allow(clientID, action, nodeID, recursive)
}

func (p *ABACPolicy) RemoveRule(clientID ClientID, action ABACAction, nodeID NodeID) {
	if actions, ok := p.Rules[clientID]; ok {
		if nodes, ok := actions[action]; ok {
			delete(nodes, nodeID)
			if len(nodes) == 0 {
				delete(actions, action)
			}
		}
		if len(actions) == 0 {
			delete(p.Rules, clientID)
		}
	}
}

func (p *ABACPolicy) IsAllowed(clientID ClientID, action ABACAction, target NodeID) bool {
	if p.tree == nil {
		panic("ABACPolicy.tree is not set")
	}

	clients := []ClientID{clientID, "*"}
	for _, c := range clients {
		if actions, ok := p.Rules[c]; ok {
			// Check exact action
			if rules, ok := actions[action]; ok {
				for nodeID, rule := range rules {
					if nodeID == "*" || nodeID == target || (rule.Recursive && p.tree.isDescendant(nodeID, target)) {
						return true
					}
				}
			}
			// Check wildcard action
			if rules, ok := actions["*"]; ok {
				for nodeID, rule := range rules {
					if nodeID == "*" || nodeID == target || (rule.Recursive && p.tree.isDescendant(nodeID, target)) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (p *ABACPolicy) MarshalJSON() ([]byte, error) {
	type Alias ABACPolicy // create an alias to avoid recursion
	return json.Marshal(&struct {
		*Alias
		Tree interface{} `json:"tree,omitempty"` // excluded from output
	}{
		Alias: (*Alias)(p),
	})
}

func (p *ABACPolicy) UnmarshalJSON(data []byte) error {
	type Alias ABACPolicy
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	return json.Unmarshal(data, &aux)
}

func (p *ABACPolicy) Hash() (string, error) {
	rulesJSON, err := json.Marshal(p.Rules)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(rulesJSON)
	return hex.EncodeToString(hash[:]), nil
}
