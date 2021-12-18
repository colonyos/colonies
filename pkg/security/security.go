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

const digestLength = 64

func GenerateCredentials(prvKey string) (string, string, string, error) {
	digest := GenerateDigest()
	sig, err := GenerateSignature(digest, prvKey)
	if err != nil {
		return "", "", "", err
	}

	id, err := GenerateID(prvKey)
	if err != nil {
		return "", "", "", err
	}

	return digest, sig, id, nil
}

func GenerateID(privateKey string) (string, error) {
	identify, err := crypto.CreateIdendityFromString(privateKey)
	if err != nil {
		return "", nil
	}

	return identify.ID(), nil
}

func GeneratePrivateKey() (string, error) {
	identify, err := crypto.CreateIdendity()
	if err != nil {
		return "", nil
	}

	return identify.PrivateKeyAsHex(), nil
}

func GenerateDigest() string {
	n := digestLength
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

	hash := crypto.GenerateHash(b)
	return hash.String()
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

func Authenticate(claimedID string, digest string, signature string) error {
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}

	hash := crypto.GenerateHashFromString(digest)
	derivedID, err := crypto.RecoveredID(hash, []byte(signatureBytes))
	if err != nil {
		return err
	}

	if derivedID != claimedID {
		return errors.New("invalid id, authentication failed")
	}

	return nil
}

func VerifyColonyOwnership(id string, colonyID string, digest string, signature string, ownership Ownership) error {
	if id != colonyID {
		return errors.New("invalid id")
	}

	err := Authenticate(id, digest, signature)
	if err != nil {
		return err
	}

	err = ownership.CheckIfColonyExists(colonyID)
	if err != nil {
		return err
	}

	return nil
}

func VerifyWorkerMembership(id string, colonyID string, digest string, signature string, ownership Ownership) error {
	err := Authenticate(id, digest, signature)
	if err != nil {
		return err
	}

	return ownership.CheckIfWorkerBelongsToColony(id, colonyID)
}

func VerifyAccessRights(id string, colonyID string, digest string, signature string, ownership Ownership) error {
	err := VerifyColonyOwnership(id, colonyID, digest, signature, ownership)
	if err != nil {
		err = VerifyWorkerMembership(id, colonyID, digest, signature, ownership)
		if err != nil {
			return err
		}

		err = ownership.CheckIfWorkerBelongsToColony(id, colonyID)
		if err != nil {
			return err
		}
	}

	return nil
}
