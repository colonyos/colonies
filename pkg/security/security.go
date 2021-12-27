package security

import (
	"colonies/pkg/crypto"
	"encoding/hex"
	"errors"
	"math/rand"
)

// TODO: Pending or disapproved runtimes should be blocked!

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

const digestLength = 64

func GenerateCredentials(prvKey string) (string, string, string, error) {
	digest := GenerateDigest()
	signature, err := GenerateSignature(digest, prvKey)
	if err != nil {
		return "", "", "", err
	}

	id, err := GenerateID(prvKey)
	if err != nil {
		return "", "", "", err
	}

	return digest, signature, id, nil
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

func RequireRoot(rootPassword string, expectedRootPassword string) error {
	if rootPassword == "" {
		return errors.New("Root password is missing")
	}

	if rootPassword != expectedRootPassword {
		return errors.New("Invalid root password")
	}

	return nil
}

func GenerateSignature(jsonString string, prvKey string) (string, error) {
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
		return errors.New("Invalid ID, authentication failed")
	}

	return nil
}

func RequireColonyOwner(id string, colonyID string, digest string, signature string, ownership Ownership) error {
	if id != colonyID {
		return errors.New("Provided ID does not match colonyID")
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

func VerifyRuntimeMembership(runtimeID string, colonyID string, ownership Ownership) error {
	return ownership.CheckIfRuntimeBelongsToColony(runtimeID, colonyID)
}

func RequireColonyMember(id string, colonyID string, digest string, signature string, ownership Ownership) error {
	err := Authenticate(id, digest, signature)
	if err != nil {
		return err
	}

	return ownership.CheckIfRuntimeBelongsToColony(id, colonyID)
}

func RequireColonyOwnerOrMember(id string, colonyID string, digest string, signature string, ownership Ownership) error {
	err := RequireColonyOwner(id, colonyID, digest, signature, ownership)
	if err != nil {
		err = RequireColonyMember(id, colonyID, digest, signature, ownership)
		if err != nil {
			return err
		}

		err = ownership.CheckIfRuntimeBelongsToColony(id, colonyID)
		if err != nil {
			return err
		}
	}

	return nil
}
