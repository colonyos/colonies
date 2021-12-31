package crypto

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

type Hash struct {
	bytes []byte
}

func GenerateHash(buf []byte) *Hash {
	hash := &Hash{}
	d := sha3.New256()
	d.Write([]byte(buf))
	hash.bytes = d.Sum(nil)

	return hash
}

func GenerateHashFromString(str string) *Hash {
	return GenerateHash([]byte(str))
}

func CreateHashFromString(str string) (*Hash, error) {
	hash := &Hash{}
	bytes, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	hash.bytes = bytes

	return hash, nil
}

func (hash *Hash) Bytes() []byte {
	return hash.bytes
}

func (hash *Hash) String() string {
	return string(hex.EncodeToString(hash.bytes))
}
