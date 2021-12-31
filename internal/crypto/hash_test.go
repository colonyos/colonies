package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateHash(t *testing.T) {
	hash := GenerateHash([]byte("hello world"))
	assert.Equal(t, "644bcc7e564373040999aac89e7622f3ca71fba1d972fd94a31c3bfbf24e3938", hash.String())

	hash = GenerateHashFromString("hello world")
	assert.Equal(t, "644bcc7e564373040999aac89e7622f3ca71fba1d972fd94a31c3bfbf24e3938", hash.String())

	byteArray := []byte{100, 75, 204, 126, 86, 67, 115, 4, 9, 153, 170, 200, 158, 118, 34, 243, 202, 113, 251, 161, 217, 114, 253, 148, 163, 28, 59, 251, 242, 78, 57, 56}
	assert.Equal(t, byteArray, hash.Bytes())
}

func TestCreateHashFromString(t *testing.T) {
	hash, err := CreateHashFromString("644bcc7e564373040999aac89e7622f3ca71fba1d972fd94a31c3bfbf24e3938")
	assert.Nil(t, err)
	assert.Equal(t, "644bcc7e564373040999aac89e7622f3ca71fba1d972fd94a31c3bfbf24e3938", hash.String())

	byteArray := []byte{100, 75, 204, 126, 86, 67, 115, 4, 9, 153, 170, 200, 158, 118, 34, 243, 202, 113, 251, 161, 217, 114, 253, 148, 163, 28, 59, 251, 242, 78, 57, 56}
	assert.Equal(t, byteArray, hash.Bytes())
}
