package crypto

import (
	"encoding/hex"

	"github.com/colonyos/colonies/internal/crypto"
)

type StandaloneCrypto struct {
}

func CreateCrypto() *StandaloneCrypto {
	return &StandaloneCrypto{}
}

func (standaloneCrypto *StandaloneCrypto) GeneratePrivateKey() (string, error) {
	identify, err := crypto.CreateIdendity()
	if err != nil {
		return "", err
	}

	return identify.PrivateKeyAsHex(), nil
}

func (standaloneCrypto *StandaloneCrypto) GenerateID(prvKey string) (string, error) {
	identify, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", err
	}

	return identify.ID(), nil
}

func (standaloneCrypto *StandaloneCrypto) GenerateSignature(data string, prvKey string) (string, error) {
	idendity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", err
	}

	hash := crypto.GenerateHashFromString(data)
	signatureBytes, err := crypto.Sign(hash, idendity.PrivateKey())
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(signatureBytes), nil
}

func (standaloneCrypto *StandaloneCrypto) GenerateHash(data string) string {
	return crypto.GenerateHash([]byte(data)).String()
}

func (standaloneCrypto *StandaloneCrypto) RecoverID(data string, signature string) (string, error) {
	signatureString, err := hex.DecodeString(signature)
	if err != nil {
		return "", err
	}

	hash := crypto.GenerateHashFromString(data)
	return crypto.RecoveredID(hash, []byte(signatureString))
}
