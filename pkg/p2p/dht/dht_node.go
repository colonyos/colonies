package dht

import (
	"github.com/colonyos/colonies/internal/crypto"
	"github.com/colonyos/colonies/pkg/p2p/libp2p"
)

func CreateDHT(addr []string) (DHT, error) {
	m, err := libp2p.CreateMessenger(addr)
	if err != nil {
		return nil, err
	}

	id, err := crypto.CreateIdendity()
	if err != nil {
		return nil, err
	}

	prvKey := id.PrivateKeyAsHex()

	contact, err := CreateContact(m.Node, prvKey)
	if err != nil {
		return nil, err
	}

	k, err := CreateKademlia(m, contact)
	if err != nil {
		return nil, err
	}

	return k, nil
}
