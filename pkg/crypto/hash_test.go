package crypto

import (
	. "colonies/pkg/utils"
	"testing"
)

func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func TestGenerateHash(t *testing.T) {
	hash := GenerateHash([]byte("hello world"))
	if hash.String() != "644bcc7e564373040999aac89e7622f3ca71fba1d972fd94a31c3bfbf24e3938" {
		Fatal(t, "invalid hash string")
	}

	hash = GenerateHashFromString("hello world")
	if hash.String() != "644bcc7e564373040999aac89e7622f3ca71fba1d972fd94a31c3bfbf24e3938" {
		Fatal(t, "invalid hash string")
	}

	byteArray := []byte{100, 75, 204, 126, 86, 67, 115, 4, 9, 153, 170, 200, 158, 118, 34, 243, 202, 113, 251, 161, 217, 114, 253, 148, 163, 28, 59, 251, 242, 78, 57, 56}
	if !Equal(byteArray, hash.Bytes()) {
		Fatal(t, "invalid hash buffer")
	}
}

func TestCreateHashFromString(t *testing.T) {
	hash, err := CreateHashFromString("644bcc7e564373040999aac89e7622f3ca71fba1d972fd94a31c3bfbf24e3938")
	CheckError(t, err)
	if hash.String() != "644bcc7e564373040999aac89e7622f3ca71fba1d972fd94a31c3bfbf24e3938" {
		Fatal(t, "invalid hash string")
	}

	byteArray := []byte{100, 75, 204, 126, 86, 67, 115, 4, 9, 153, 170, 200, 158, 118, 34, 243, 202, 113, 251, 161, 217, 114, 253, 148, 163, 28, 59, 251, 242, 78, 57, 56}
	if !Equal(byteArray, hash.Bytes()) {
		Fatal(t, "invalid hash buffer")
	}
}
