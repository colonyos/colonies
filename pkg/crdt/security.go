package crdt

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/colonyos/colonies/internal/crypto"
	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
)

type nodeDigest struct {
	ID           NodeID         `json:"id"`
	ParentID     NodeID         `json:"parentid"`
	Edges        []*EdgeCRDT    `json:"edges"`
	Clock        map[string]int `json:"clock"`
	Owner        ClientID       `json:"owner"`
	IsRoot       bool           `json:"isroot"`
	IsMap        bool           `json:"ismap"`
	IsArray      bool           `json:"isarray"`
	IsLiteral    bool           `json:"isliteral"`
	LiteralValue interface{}    `json:"litteralValue"`
	Nounce       string         `json:"nounce"`
	IsDeleted    bool           `json:"deleted"`
}

// We cannot calculate the digest edges and clock here because they will change after a merge operation
func (n *NodeCRDT) ComputeDigest() (*crypto.Hash, error) {
	d := nodeDigest{
		ID:           n.ID,
		Owner:        n.Owner,
		IsRoot:       n.IsRoot,
		IsMap:        n.IsMap,
		IsArray:      n.IsArray,
		IsLiteral:    n.IsLiteral,
		LiteralValue: n.LiteralValue,
		Nounce:       n.Nounce,
		IsDeleted:    n.IsDeleted,
	}

	var buf bytes.Buffer
	buf.WriteString("{")

	encodeField(&buf, "id", d.ID)
	encodeField(&buf, "parentid", d.ParentID)
	encodeField(&buf, "edges", d.Edges)

	err := encodeSortedMap(&buf, "clock", d.Clock)
	if err != nil {
		return nil, err
	}

	encodeField(&buf, "owner", d.Owner)
	encodeField(&buf, "isroot", d.IsRoot)
	encodeField(&buf, "ismap", d.IsMap)
	encodeField(&buf, "isarray", d.IsArray)
	encodeField(&buf, "isliteral", d.IsLiteral)
	encodeField(&buf, "litteralValue", d.LiteralValue)
	encodeField(&buf, "nounce", d.Nounce)
	encodeField(&buf, "deleted", d.IsDeleted)

	buf.Truncate(buf.Len() - 1) // remove last comma
	buf.WriteString("}")

	digest := crypto.GenerateHashFromString(string(buf.Bytes()) + n.Nounce)

	return digest, nil
}

// Helper to encode one field
func encodeField(buf *bytes.Buffer, name string, value interface{}) {
	b, _ := json.Marshal(value)
	buf.WriteString(`"`)
	buf.WriteString(name)
	buf.WriteString(`":`)
	buf.Write(b)
	buf.WriteString(",")
}

// Helper to encode sorted map
func encodeSortedMap(buf *bytes.Buffer, name string, m map[string]int) error {
	buf.WriteString(`"`)
	buf.WriteString(name)
	buf.WriteString(`":{`)

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		vString, err := jsonInt(v)
		if err != nil {
			return err
		}
		buf.WriteString(`"`)
		buf.WriteString(k)
		buf.WriteString(`":`)
		buf.WriteString(vString)
		buf.WriteString(",")
	}

	if len(keys) > 0 {
		buf.Truncate(buf.Len() - 1) // remove last comma
	}

	buf.WriteString("},")
	return nil
}

func jsonInt(i int) (string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (n *NodeCRDT) Sign(identity *crypto.Idendity) error {
	n.Nounce = core.GenerateRandomID()
	digest, err := n.ComputeDigest()
	if err != nil {
		log.WithFields(log.Fields{
			"NodeID": n.ID,
			"Error":  err,
		}).Error("Failed to compute node digest")
		return fmt.Errorf("Failed to compute node digest: %w", err)
	}

	signature, err := crypto.Sign(digest, identity.PrivateKey())
	if err != nil {
		log.WithFields(log.Fields{
			"NodeID": n.ID,
			"Error":  err,
		}).Error("Failed to sign node")
		return fmt.Errorf("Failed to sign node: %w", err)
	}

	signatureStr := hex.EncodeToString(signature)
	n.Signature = signatureStr

	return nil
}

func (n *NodeCRDT) Verify() (string, error) {
	digest, err := n.ComputeDigest()
	if err != nil {
		log.WithFields(log.Fields{
			"NodeID": n.ID,
			"Error":  err,
		}).Error("Failed to compute node digest")
		return "", fmt.Errorf("Failed to compute node digest: %w", err)
	}
	signatureBytes, err := hex.DecodeString(n.Signature)
	if err != nil {
		log.WithFields(log.Fields{
			"NodeID": n.ID,
			"Error":  err,
		}).Error("Failed to decode signature")
		return "", fmt.Errorf("Failed to decode signature: %w", err)
	}

	recoveredPublicKey, err := crypto.RecoverPublicKey(digest, signatureBytes)
	if err != nil {
		log.WithFields(log.Fields{
			"NodeID": n.ID,
			"Error":  err,
		}).Error("Failed to recover public key from signature")
		return "", fmt.Errorf("Failed to recover public key from signature: %w", err)
	}

	valid, err := crypto.Verify(recoveredPublicKey, digest, signatureBytes)
	if err != nil {
		log.WithFields(log.Fields{
			"NodeID": n.ID,
			"Error":  err,
		}).Error("Failed to verify signature")
		return "", fmt.Errorf("Failed to verify signature: %w", err)
	}
	if !valid {
		log.WithFields(log.Fields{
			"NodeID":    n.ID,
			"Signature": n.Signature,
			"Digest":    digest,
		}).Error("Signature verification failed")
		return "", fmt.Errorf("Signature verification failed for node %s", n.ID)
	}

	recoveredID, err := crypto.RecoveredID(digest, signatureBytes)
	if err != nil {
		log.WithFields(log.Fields{
			"NodeID": n.ID,
			"Error":  err,
		}).Error("Failed to recover ID from signature")
		return "", fmt.Errorf("Failed to recover ID from signature: %w", err)
	}

	if recoveredID != string(n.Owner) {
		log.WithFields(log.Fields{
			"NodeID":        n.ID,
			"RecoveredID":   recoveredID,
			"ExpectedOwner": n.Owner,
		}).Error("Recovered ID does not match node owner")
		return "", fmt.Errorf("Recovered ID %s does not match node owner %s", recoveredID, n.Owner)
	}

	return recoveredID, nil
}
