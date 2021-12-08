package core

import (
	"colonies/pkg/crypto"
)

type Colony struct {
	name     string
	identity *crypto.Idendity
}

func CreateColony(name string) (*Colony, error) {
	colony := &Colony{name: name}

	identity, err := crypto.CreateIdendity()
	if err != nil {
		return nil, err
	}

	colony.identity = identity

	return colony, nil
}

func CreateColonyFromDB(name string, privateKey string) (*Colony, error) {
	colony := &Colony{name: name}
	var err error
	colony.identity, err = crypto.CreateIdendityFromString(privateKey)
	if err != nil {
		return nil, err
	}

	return colony, nil
}

func (colony *Colony) Name() string {
	return colony.name
}

func (colony *Colony) ID() string {
	return colony.identity.ID()
}

func (colony *Colony) PrivateKey() string {
	return colony.identity.PrivateKeyAsHex()
}
