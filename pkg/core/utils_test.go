package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomID(t *testing.T) {
	str := GenerateRandomID()
	assert.Len(t, str, 64)
}
