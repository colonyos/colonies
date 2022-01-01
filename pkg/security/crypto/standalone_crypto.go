package crypto

import (
	"colonies/internal/crypto"
	"encoding/hex"
)

type StandaloneCrypto struct {
}

func CreateCrypto() *StandaloneCrypto {
	return &StandaloneCrypto{}
}

func (standaloneCrypto *StandaloneCrypto) GeneratePrivateKey() (string, error) {
	identify, err := crypto.CreateIdendity()
	if err != nil {
		return "", nil
	}

	return identify.PrivateKeyAsHex(), nil
}

func (standaloneCrypto *StandaloneCrypto) GenerateID(prvKey string) (string, error) {
	identify, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", nil
	}

	return identify.ID(), nil
}

func (standaloneCrypto *StandaloneCrypto) GenerateSignature(jsonString string, prvKey string) (string, error) {
	idendity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", err
	}

	hash := crypto.GenerateHashFromString(jsonString)
	signatureBytes, err := crypto.Sign(hash, idendity.PrivateKey())
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(signatureBytes), nil
}

func (standaloneCrypto *StandaloneCrypto) GenerateHash(data string) string {
	return crypto.GenerateHash([]byte(data)).String()
}

func (standaloneCrypto *StandaloneCrypto) RecoverID(jsonString string, signature string) (string, error) {
	signatureString, err := hex.DecodeString(signature)
	if err != nil {
		return "", err
	}

	hash := crypto.GenerateHashFromString(jsonString)
	return crypto.RecoveredID(hash, []byte(signatureString))
}
