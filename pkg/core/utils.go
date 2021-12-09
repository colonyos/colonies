package core

import (
	"colonies/pkg/crypto"

	"github.com/google/uuid"
)

func GenerateRandomID() string {
	uuid := uuid.New()
	return crypto.GenerateHashFromString(uuid.String()).String()
}
