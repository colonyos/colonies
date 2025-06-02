package crdt

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/colonyos/colonies/internal/crypto"
	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
)

type ABACAction string

const (
	ActionModify ABACAction = "modify"
	ActionRead   ABACAction = "read"
)

type ABACRule struct {
	Recursive bool `json:"recursive"`
}

type TreeChecker interface {
	isDescendant(root NodeID, target NodeID) bool
}

type ABACPolicy struct {
	Rules     map[string]map[ABACAction]map[NodeID]ABACRule `json:"rules"`
	OwnerID   string                                        `json:"ownerID"`
	Clock     VectorClock                                   `json:"clock"`
	Nounce    string                                        `json:"nounce"`
	Signature string                                        `json:"signature"`
	tree      TreeChecker                                   `json:"-"`
	identity  *crypto.Idendity                              `json:"-"`
}

func NewABACPolicy(tree TreeChecker, ownerID string, identity *crypto.Idendity) *ABACPolicy {
	return &ABACPolicy{
		Rules:    make(map[string]map[ABACAction]map[NodeID]ABACRule),
		tree:     tree,
		OwnerID:  ownerID,
		identity: identity,
		Clock:    make(VectorClock),
	}
}

func (p *ABACPolicy) SetTree(tree TreeChecker) {
	p.tree = tree
}

func (p *ABACPolicy) Allow(id string, action ABACAction, nodeID NodeID, recursive bool) error {
	clientID := ClientID(p.identity.ID())

	// Step 1: Prepare new clock
	newClock := copyClock(p.Clock)

	maxVersion := 0
	for _, v := range newClock {
		if v > maxVersion {
			maxVersion = v
		}
	}
	version := maxVersion + 1
	newClock[clientID] = version

	// Step 2: Resolve conflict
	winningClock, winningOwner := resolveConflict(p.Clock, newClock, ClientID(p.OwnerID), clientID, false)

	if clocksEqual(winningClock, newClock) && winningOwner == clientID {
		// Step 3: Apply update
		p.Clock = newClock

		if _, ok := p.Rules[id]; !ok {
			p.Rules[id] = make(map[ABACAction]map[NodeID]ABACRule)
		}
		if _, ok := p.Rules[id][action]; !ok {
			p.Rules[id][action] = make(map[NodeID]ABACRule)
		}
		p.Rules[id][action][nodeID] = ABACRule{Recursive: recursive}

		// Step 4: Sign
		err := p.Sign()
		if err != nil {
			log.WithFields(log.Fields{
				"OwnerID": p.OwnerID,
				"Error":   err,
			}).Error("Failed to sign ABACPolicy after allowing rule")
			return fmt.Errorf("Failed to sign ABACPolicy after allowing rule: %w", err)
		}
	} else {
		// Step 5: Conflict — ignore update
		log.WithFields(log.Fields{
			"OwnerID":      p.OwnerID,
			"ClientID":     clientID,
			"WinningOwner": winningOwner,
			"WinningClock": winningClock,
			"NewClock":     newClock,
		}).Debug("ABACPolicy Allow ignored due to conflict resolution")
		// Do not return an error — just ignore
	}

	return nil
}

func (p *ABACPolicy) UpdateRule(id string, action ABACAction, nodeID NodeID, recursive bool) error {
	return p.Allow(id, action, nodeID, recursive)
}

func (p *ABACPolicy) RemoveRule(id string, action ABACAction, nodeID NodeID) error {
	clientID := ClientID(p.identity.ID())

	// Step 1: Prepare new clock
	newClock := copyClock(p.Clock)

	maxVersion := 0
	for _, v := range newClock {
		if v > maxVersion {
			maxVersion = v
		}
	}
	version := maxVersion + 1
	newClock[clientID] = version

	// Step 2: Resolve conflict
	winningClock, winningOwner := resolveConflict(p.Clock, newClock, ClientID(p.OwnerID), clientID, false)

	if clocksEqual(winningClock, newClock) && winningOwner == clientID {
		// Step 3: Apply update
		p.Clock = newClock

		if actions, ok := p.Rules[id]; ok {
			if nodes, ok := actions[action]; ok {
				delete(nodes, nodeID)
				if len(nodes) == 0 {
					delete(actions, action)
				}
			}
			if len(actions) == 0 {
				delete(p.Rules, id)
			}
		}

		// Step 4: Sign
		err := p.Sign()
		if err != nil {
			log.WithFields(log.Fields{
				"OwnerID": p.OwnerID,
				"Error":   err,
			}).Error("Failed to sign ABACPolicy after removing rule")
			return fmt.Errorf("Failed to sign ABACPolicy after removing rule: %w", err)
		}
	} else {
		// Step 5: Conflict — ignore update
		log.WithFields(log.Fields{
			"OwnerID":      p.OwnerID,
			"ClientID":     clientID,
			"WinningOwner": winningOwner,
			"WinningClock": winningClock,
			"NewClock":     newClock,
		}).Debug("ABACPolicy RemoveRule ignored due to conflict resolution")
	}

	return nil
}

func (p *ABACPolicy) IsAllowed(id string, action ABACAction, target NodeID) bool {
	if p.tree == nil {
		panic("ABACPolicy.tree is not set")
	}

	recoveredID, err := p.Verify()
	if err != nil {
		log.WithFields(log.Fields{
			"OwnerID":     p.OwnerID,
			"RecoveredID": recoveredID,
			"Error":       err,
		}).Error("ABACPolicy verification failed, recovered ID does not match owner ID or signature verification failed")
		return false
	}

	clients := []string{id, "*"}
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

func (p *ABACPolicy) Merge(remote *ABACPolicy) error {
	// Step 1: Resolve policy-level conflict (LWW on full policy)
	winningClock, winningOwner := resolveConflict(
		p.Clock,
		remote.Clock,
		ClientID(p.OwnerID),
		ClientID(remote.OwnerID),
		false, // not append → LWW mode
	)

	if clocksEqual(winningClock, remote.Clock) && winningOwner == ClientID(remote.OwnerID) {
		// Remote wins → replace entire ABACPolicy
		p.Clock = copyClock(remote.Clock)
		p.Rules = deepCopyRules(remote.Rules)
		p.OwnerID = remote.OwnerID
		p.Nounce = remote.Nounce
		p.Signature = remote.Signature

		log.WithFields(log.Fields{
			"LocalOwner":  p.OwnerID,
			"RemoteOwner": remote.OwnerID,
			"Winning":     "remote",
		}).Info("ABACPolicy Merge: remote policy wins, replaced local policy")
	} else {
		// Local wins → do nothing
		log.WithFields(log.Fields{
			"LocalOwner":  p.OwnerID,
			"RemoteOwner": remote.OwnerID,
			"Winning":     "local",
		}).Info("ABACPolicy Merge: local policy wins, no changes applied")
	}

	return nil
}

func deepCopyRules(rules map[string]map[ABACAction]map[NodeID]ABACRule) map[string]map[ABACAction]map[NodeID]ABACRule {
	newRules := make(map[string]map[ABACAction]map[NodeID]ABACRule)
	for clientID, actions := range rules {
		newActions := make(map[ABACAction]map[NodeID]ABACRule)
		for action, nodes := range actions {
			newNodes := make(map[NodeID]ABACRule)
			for nodeID, rule := range nodes {
				newNodes[nodeID] = rule
			}
			newActions[action] = newNodes
		}
		newRules[clientID] = newActions
	}
	return newRules
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

func (p *ABACPolicy) ComputeDigest() (*crypto.Hash, error) {
	// Build an ordered structure for Rules
	orderedRules := make([]struct {
		ClientID  string
		Action    string
		NodeID    string
		Recursive bool
	}, 0)

	// Sort client IDs
	clientIDs := make([]string, 0, len(p.Rules))
	for clientID := range p.Rules {
		clientIDs = append(clientIDs, clientID)
	}
	sort.Strings(clientIDs)

	for _, clientID := range clientIDs {
		actions := p.Rules[clientID]

		// Sort actions
		actionKeys := make([]string, 0, len(actions))
		for action := range actions {
			actionKeys = append(actionKeys, string(action))
		}
		sort.Strings(actionKeys)

		for _, actionStr := range actionKeys {
			action := ABACAction(actionStr)
			rules := actions[action]

			// Sort node IDs
			nodeIDs := make([]string, 0, len(rules))
			for nodeID := range rules {
				nodeIDs = append(nodeIDs, string(nodeID))
			}
			sort.Strings(nodeIDs)

			for _, nodeID := range nodeIDs {
				rule := rules[NodeID(nodeID)]
				orderedRules = append(orderedRules, struct {
					ClientID  string
					Action    string
					NodeID    string
					Recursive bool
				}{
					ClientID:  clientID,
					Action:    actionStr,
					NodeID:    nodeID,
					Recursive: rule.Recursive,
				})
			}
		}
	}

	// Now build the full digest input struct
	digestInput := struct {
		Rules   interface{} `json:"rules"`
		OwnerID string      `json:"ownerID"`
		Nounce  string      `json:"nounce"`
	}{
		Rules:   orderedRules,
		OwnerID: p.OwnerID,
		Nounce:  p.Nounce,
	}

	// Marshal the entire digest input
	buf, err := json.Marshal(digestInput)
	if err != nil {
		return nil, fmt.Errorf("ComputeDigest: failed to marshal digest input: %w", err)
	}

	// Now generate the hash from the full JSON string
	digest := crypto.GenerateHashFromString(string(buf))

	return digest, nil
}

func (p *ABACPolicy) PrintPolicy() {
	fmt.Println("ABAC Policy:")
	fmt.Println("============")

	if p.Rules == nil || len(p.Rules) == 0 {
		fmt.Println("(empty)")
		return
	}

	// Sort clients for stable output
	clientIDs := make([]string, 0, len(p.Rules))
	for clientID := range p.Rules {
		clientIDs = append(clientIDs, clientID)
	}
	sort.Strings(clientIDs)

	for _, clientID := range clientIDs {
		fmt.Printf("Client: %s\n", clientID)

		// Sort actions for stable output
		actions := p.Rules[clientID]
		actionKeys := make([]string, 0, len(actions))
		for action := range actions {
			actionKeys = append(actionKeys, string(action))
		}
		sort.Strings(actionKeys)

		for _, actionStr := range actionKeys {
			action := ABACAction(actionStr)
			fmt.Printf("  Action: %s\n", action)

			// Sort nodeIDs for stable output
			rules := actions[action]
			nodeIDs := make([]string, 0, len(rules))
			for nodeID := range rules {
				nodeIDs = append(nodeIDs, string(nodeID))
			}
			sort.Strings(nodeIDs)

			for _, nodeID := range nodeIDs {
				rule := rules[NodeID(nodeID)]
				fmt.Printf("    Node: %s (Recursive: %v)\n", nodeID, rule.Recursive)
			}
		}
	}

	fmt.Println()
}

func (p *ABACPolicy) Sign() error {
	p.Nounce = core.GenerateRandomID()

	digest, err := p.ComputeDigest()
	if err != nil {
		log.WithFields(log.Fields{
			"OwnerID": p.OwnerID,
			"Error":   err,
		}).Error("Failed to compute ABACPolicy digest")
		return fmt.Errorf("Failed to compute ABACPolicy digest: %w", err)
	}

	signature, err := crypto.Sign(digest, p.identity.PrivateKey())
	if err != nil {
		log.WithFields(log.Fields{
			"OwnerID": p.OwnerID,
			"Error":   err,
		}).Error("Failed to sign ABACPolicy")
		return fmt.Errorf("Failed to sign ABACPolicy: %w", err)
	}

	p.Signature = hex.EncodeToString(signature)

	return nil
}

func (p *ABACPolicy) Verify() (string, error) {
	digest, err := p.ComputeDigest()
	if err != nil {
		log.WithFields(log.Fields{
			"OwnerID": p.OwnerID,
			"Error":   err,
		}).Error("Failed to compute ABACPolicy digest")
		return "", fmt.Errorf("Failed to compute ABACPolicy digest: %w", err)
	}

	signatureBytes, err := hex.DecodeString(p.Signature)
	if err != nil {
		log.WithFields(log.Fields{
			"OwnerID": p.OwnerID,
			"Error":   err,
		}).Error("Failed to decode ABACPolicy signature")
		return "", fmt.Errorf("Failed to decode ABACPolicy signature: %w", err)
	}

	recoveredPublicKey, err := crypto.RecoverPublicKey(digest, signatureBytes)
	if err != nil {
		log.WithFields(log.Fields{
			"OwnerID": p.OwnerID,
			"Error":   err,
		}).Error("Failed to recover public key from ABACPolicy signature")
		return "", fmt.Errorf("Failed to recover public key from signature: %w", err)
	}

	valid, err := crypto.Verify(recoveredPublicKey, digest, signatureBytes)
	if err != nil {
		log.WithFields(log.Fields{
			"OwnerID": p.OwnerID,
			"Error":   err,
		}).Error("Failed to verify ABACPolicy signature")
		return "", fmt.Errorf("Failed to verify signature: %w", err)
	}

	if !valid {
		log.WithFields(log.Fields{
			"OwnerID":   p.OwnerID,
			"Signature": p.Signature,
			"Digest":    digest,
		}).Error("ABACPolicy signature verification failed")
		return "", fmt.Errorf("Signature verification failed for ABACPolicy owned by %s", p.OwnerID)
	}

	recoveredID, err := crypto.RecoveredID(digest, signatureBytes)
	if err != nil {
		log.WithFields(log.Fields{
			"OwnerID": p.OwnerID,
			"Error":   err,
		}).Error("Failed to recover ID from ABACPolicy signature")
		return "", fmt.Errorf("Failed to recover ID from signature: %w", err)
	}

	if recoveredID != p.OwnerID {
		log.WithFields(log.Fields{
			"OwnerID":     p.OwnerID,
			"RecoveredID": recoveredID,
		}).Error("Recovered ID does not match ABACPolicy owner")
		return "", fmt.Errorf("Recovered ID %s does not match ABACPolicy owner %s", recoveredID, p.OwnerID)
	}

	return recoveredID, nil
}

func (p *ABACPolicy) Clone() (*ABACPolicy, error) {
	j, err := json.Marshal(p)
	if err != nil {
		log.WithFields(log.Fields{
			"OwnerID": p.OwnerID,
			"Error":   err,
		}).Error("Failed to marshal ABACPolicy for cloning")
		return nil, fmt.Errorf("Failed to marshal ABACPolicy for cloning: %w", err)
	}
	clone := &ABACPolicy{}
	err = json.Unmarshal(j, clone)
	if err != nil {
		log.WithFields(log.Fields{
			"OwnerID": p.OwnerID,
			"Error":   err,
		}).Error("Failed to unmarshal ABACPolicy for cloning")
		return nil, fmt.Errorf("Failed to unmarshal ABACPolicy for cloning: %w", err)
	}

	return clone, nil
}
