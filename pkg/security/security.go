package security

import (
	"colonies/pkg/crypto"
	"encoding/hex"
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckAPIKey(c *gin.Context, expectedAPIKey string) error {
	apiKey := c.GetHeader("Api-Key")
	if apiKey == "" {
		return errors.New("Api-Key header not specified")
	}

	if apiKey != expectedAPIKey {
		return errors.New("Invalid Api-Key")
	}

	return nil
}

func GenSignature(jsonString string, prvKey string) (string, error) {
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
