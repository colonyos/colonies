package security

import (
	"colonies/pkg/crypto"
	"encoding/hex"
	"errors"
	"math/rand"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func GenerateRandomString() string {
	n := 64
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func VerifyAPIKey(apiKey string, expectedAPIKey string) error {
	if apiKey == "" {
		return errors.New("Api-Key is missing")
	}

	if apiKey != expectedAPIKey {
		return errors.New("Invalid Api-Key")
	}

	return nil
}

func GenerateSignature(jsonString string, prvKey string) (string, error) {
	idendity, err := crypto.CreateIdendityFromString(prvKey)
	if err != nil {
		return "", err
	}

	hash := crypto.GenerateHashFromString(jsonString)
	sig, err := crypto.Sign(hash, idendity.PrivateKey())
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(sig), nil
}

func VerifyColonyOwnership(colonyID string, data string, signature string, ownership Ownership) error {
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}

	hash := crypto.GenerateHash([]byte(data))
	derivedColonyID, err := crypto.RecoveredID(hash, []byte(signatureBytes))
	if err != nil {
		return err
	}

	if derivedColonyID != colonyID {
		return errors.New("invalid signature")
	}

	err = ownership.CheckIfColonyExists(colonyID)
	if err != nil {
		return err
	}

	return nil
}
