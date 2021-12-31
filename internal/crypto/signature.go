package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/pkg/errors"
)

var (
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
	wordBits       = 32 << (uint64(^big.Word(0)) >> 63)
	wordBytes      = wordBits / 8
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func RecoveredID(hash *Hash, sig []byte) (string, error) {
	if len(sig) != 65 {
		return "", errors.New("Invalid signature length")
	}

	if len(hash.Bytes()) != 32 {
		return "", errors.New("Invalid hash length")
	}

	pub, err := RecoverPublicKey(hash, sig)
	if err != nil {
		return "", errors.Wrap(err, "Failed to recover public key")
	}

	validSig, err := Verify(pub, hash, sig)
	if err != nil {
		return "", errors.Wrap(err, "Failed verify signature")
	}
	if !validSig {
		return "", errors.New("Invalid signature")
	}

	return GenerateHashFromString(hex.EncodeToString(pub)).String(), nil
}

func Sign(hash *Hash, prv *ecdsa.PrivateKey) ([]byte, error) {
	if len(hash.Bytes()) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes")
	}
	if prv.Curve != btcec.S256() {
		return nil, fmt.Errorf("private key curve is not secp256k1")
	}
	sig, err := btcec.SignCompact(btcec.S256(), (*btcec.PrivateKey)(prv), hash.Bytes(), false)
	if err != nil {
		return nil, err
	}

	v := sig[0] - 27
	copy(sig, sig[1:])
	sig[64] = v
	return sig, nil
}

func Verify(pubkey []byte, hash *Hash, sigrec []byte) (bool, error) {
	if len(sigrec) != 65 {
		return false, nil
	}
	if len(hash.Bytes()) != 32 {
		return false, errors.New("Invalid hash length, expected it to be 32")
	}

	sig := sigrec[:len(sigrec)-1]

	signature := &btcec.Signature{R: new(big.Int).SetBytes(sig[:32]), S: new(big.Int).SetBytes(sig[32:])}
	key, err := btcec.ParsePubKey(pubkey, btcec.S256())
	if err != nil {
		return false, nil
	}
	if signature.S.Cmp(secp256k1halfN) > 0 {
		return false, nil
	}

	return signature.Verify(hash.Bytes(), key), nil
}

func RecoverPublicKey(hash *Hash, sig []byte) ([]byte, error) {
	btcsig := make([]byte, 65)
	btcsig[0] = sig[64] + 27
	copy(btcsig[1:], sig)

	pub, _, err := btcec.RecoverCompact(btcec.S256(), btcsig, hash.Bytes())
	if err != nil {
		return nil, err
	}

	bytes := (*btcec.PublicKey)(pub).SerializeUncompressed()
	return bytes, err
}
