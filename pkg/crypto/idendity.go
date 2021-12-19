package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"time"

	"github.com/btcsuite/btcd/btcec"
)

type Idendity struct {
	prv *ecdsa.PrivateKey
	id  string
}

func CreateIdendity() (*Idendity, error) {
	mathrand.Seed(time.Now().UnixNano())
	idendity := &Idendity{}

	prv, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	idendity.prv = prv
	idendity.id = GenerateHashFromString(idendity.PublicKeyAsHex()).String()

	return idendity, nil
}

func (idendity *Idendity) PrivateKey() *ecdsa.PrivateKey {
	return idendity.prv
}

func (idendity *Idendity) PrivateKeyAsHex() string {
	n := idendity.prv.Params().BitSize / 8
	binaryDump := make([]byte, n)
	if idendity.prv.D.BitLen()/8 >= n {
		binaryDump = idendity.prv.D.Bytes()
	} else {
		i := len(binaryDump)
		for _, d := range idendity.prv.D.Bits() {
			for j := 0; j < wordBytes && i > 0; j++ {
				i--
				binaryDump[i] = byte(d)
				d >>= 8
			}
		}
	}

	return hex.EncodeToString(binaryDump)
}

func (idendity *Idendity) ID() string {
	return idendity.id
}

func CreateIdendityFromString(hexEncodedPrv string) (*Idendity, error) {
	idendity := &Idendity{}
	decodedPrv, err := hex.DecodeString(hexEncodedPrv)

	if err != nil {
		return nil, err
	}

	prv := new(ecdsa.PrivateKey)
	prv.PublicKey.Curve = btcec.S256()

	if 8*len(decodedPrv) != prv.Params().BitSize {
		return nil, fmt.Errorf("Invalid private key length, should be %d bits", prv.Params().BitSize)
	}

	prv.D = new(big.Int).SetBytes(decodedPrv)
	if prv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("Invalid private key")
	}
	if prv.D.Sign() <= 0 {
		return nil, fmt.Errorf("Invalid private key")
	}

	prv.PublicKey.X, prv.PublicKey.Y = prv.PublicKey.Curve.ScalarBaseMult(decodedPrv)
	if prv.PublicKey.X == nil {
		return nil, errors.New("Invalid private key")
	}

	idendity.prv = prv
	idendity.id = GenerateHashFromString(idendity.PublicKeyAsHex()).String()

	return idendity, nil
}

func (idendity *Idendity) PublicKey() []byte {
	return elliptic.Marshal(btcec.S256(), idendity.prv.PublicKey.X, idendity.prv.PublicKey.Y)
}

func (idendity *Idendity) PublicKeyAsHex() string {
	return hex.EncodeToString(idendity.PublicKey())
}
