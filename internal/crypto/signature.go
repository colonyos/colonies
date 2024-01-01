package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
	becdsa "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/pkg/errors"
)

var (
	secp256k1N, _  = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	secp256k1halfN = new(big.Int).Div(secp256k1N, big.NewInt(2))
	wordBits       = 32 << (uint64(^big.Word(0)) >> 63)
	wordBytes      = wordBits / 8
)

const SignatureLength = 64 + 1
const RecoveryIDOffset = 64

func RecoveredID(hash *Hash, sig []byte) (string, error) {
	if len(sig) != SignatureLength {
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
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash.Bytes()))
	}
	if prv.Curve != btcec.S256() {
		return nil, errors.New("private key curve is not secp256k1")
	}

	// Convert ecdsa.PrivateKey -> btcec.PrivateKey
	var priv btcec.PrivateKey
	if overflow := priv.Key.SetByteSlice(prv.D.Bytes()); overflow || priv.Key.IsZero() {
		return nil, errors.New("invalid private key")
	}
	defer priv.Zero()
	sig, err := becdsa.SignCompact(&priv, hash.Bytes(), false) // ref uncompressed pubkey
	if err != nil {
		return nil, err
	}

	v := sig[0] - 27
	copy(sig, sig[1:])
	sig[RecoveryIDOffset] = v

	return sig, nil
}

func Verify(pubkey []byte, hash *Hash, sig []byte) (bool, error) {
	sig = sig[:len(sig)-1] // Remove recovery id

	if len(sig) != SignatureLength-1 {
		return false, fmt.Errorf("Invalid signature length")
	}
	var r, s btcec.ModNScalar
	if r.SetByteSlice(sig[:32]) {
		return false, fmt.Errorf("Failed to parse signature")
	}
	if s.SetByteSlice(sig[32:]) {
		return false, fmt.Errorf("Failed to parse signature")
	}
	signature := becdsa.NewSignature(&r, &s)
	key, err := btcec.ParsePubKey(pubkey)
	if err != nil {
		return false, err
	}

	// Reject malleable signatures, libsecp256k1 does this check but btcec doesn't
	if s.IsOverHalfOrder() {
		return false, fmt.Errorf("Signature s value is over half the order")
	}

	return signature.Verify(hash.Bytes(), key), nil
}

func RecoverPublicKey(hash *Hash, sig []byte) ([]byte, error) {
	if len(sig) != SignatureLength {
		return nil, errors.New("Invalid signature")
	}

	btcsig := make([]byte, SignatureLength)
	btcsig[0] = sig[RecoveryIDOffset] + 27
	copy(btcsig[1:], sig)

	pub, _, err := becdsa.RecoverCompact(btcsig, hash.Bytes())
	if err != nil {
		return nil, err
	}
	bytes := pub.SerializeUncompressed()
	return bytes, err
}
