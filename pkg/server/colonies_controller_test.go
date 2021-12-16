package server

import (
	"colonies/pkg/database"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateColonyController(t *testing.T) {
	db, err := database.PrepareTests()
	assert.Nil(t, err)

	CreateColoniesController(db)
}
