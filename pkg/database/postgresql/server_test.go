package postgresql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerID(t *testing.T) {
	db, err := PrepareTests()
	assert.Nil(t, err)

	err = db.SetServerID("", "server_id")
	assert.Nil(t, err)

	serverID, err := db.GetServerID()
	assert.Nil(t, err)
	assert.Equal(t, "server_id", serverID)

	err = db.SetServerID("server_id", "new_server_id")
	assert.Nil(t, err)

	serverID, err = db.GetServerID()
	assert.Nil(t, err)
	assert.Equal(t, "new_server_id", serverID)

	err = db.SetServerID("", "new_server_id")
	assert.NotNil(t, err)

	defer db.Close()
}
