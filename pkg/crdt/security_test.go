package crdt

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/colonyos/colonies/internal/crypto"
	"github.com/stretchr/testify/assert"
)

func TestInterop(t *testing.T) {
	prvKey := "d6eb959e9aec2e6fdc44b5862b269e987b8a4d6f2baca542d8acaa97ee5e74f6"
	idendity, err := crypto.CreateIdendityFromString(prvKey)
	assert.Nil(t, err)

	fmt.Println("Message=hello")
	hash := crypto.GenerateHashFromString("hello")

	signature, err := crypto.Sign(hash, idendity.PrivateKey())
	assert.Nil(t, err)
	signatureStr := hex.EncodeToString(signature)

	fmt.Println("prvkey: " + idendity.PrivateKeyAsHex())
	fmt.Println("pubkey: " + idendity.PublicKeyAsHex())
	fmt.Println("id: " + idendity.ID())
	fmt.Println("digest: " + hash.String())
	fmt.Println("signature: " + string(signatureStr))

	signaturesHex := "e713a1bb015fecabb5a084b0fe6d6e7271fca6f79525a634183cfdb175fe69241f4da161779d8e6b761200e1cf93766010a19072fa778f9643363e2cfadd640900"

	signatureBytes, err := hex.DecodeString(signaturesHex)
	recoveredID, err := crypto.RecoveredID(hash, signatureBytes)
	fmt.Println("recoveredID: " + recoveredID)
}
