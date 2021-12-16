package server

import (
	"colonies/pkg/database/postgresql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateColonyController(t *testing.T) {
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	CreateColoniesController(db)
}
