package core

import (
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/google/uuid"
)

type GeneratorArg struct {
	ID          string `json:"id"`
	GeneratorID string `json:"generatorid"`
	ColonyName  string `json:"colonyname"`
	Arg         string `json:"arg"`
}

func CreateGeneratorArg(generatorID string, colonyName string, arg string) *GeneratorArg {
	uuid := uuid.New()
	crypto := crypto.CreateCrypto()
	id := crypto.GenerateHash(uuid.String())

	generatorArg := &GeneratorArg{
		ID:          id,
		GeneratorID: generatorID,
		ColonyName:  colonyName,
		Arg:         arg,
	}

	return generatorArg
}
