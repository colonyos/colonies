package postgresql

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createDatabase() *PQDatabase {
	dbHost := os.Getenv("COLONIES_DB_HOST")
	dbPort := 5432
	dbUser := os.Getenv("COLONIES_DB_USER")
	dbPassword := os.Getenv("COLONIES_DB_PASSWORD")
	dbName := "postgres"
	dbPrefix := "TEST_"

	return CreatePQDatabase(dbHost, dbPort, dbUser, dbPassword, dbName, dbPrefix, false)
}

func TestLockClosedDB(t *testing.T) {
	db := createDatabase()

	err := db.Connect()
	assert.Nil(t, err)

	db.Close()

	err = db.Lock(1000)
	assert.NotNil(t, err)

	err = db.Unlock()
	assert.NotNil(t, err)
}

func TestLock(t *testing.T) {
	db := createDatabase()

	err := db.Connect()
	assert.Nil(t, err)

	defer db.Close()

	db.Drop()

	err = db.Initialize()

	go func() {
		time.Sleep(1 * time.Second)
		err := db.Unlock()
		assert.Nil(t, err)
	}()

	err = db.Lock(10000)
	assert.Nil(t, err)

	db2 := createDatabase()

	err = db2.Connect()
	assert.Nil(t, err)
	defer db2.Close()

	// The function below will block until db.Unlock() is called in the go-routine above
	err = db2.Lock(10000)
	assert.Nil(t, err)
}

func TestLockClose(t *testing.T) {
	db := createDatabase()

	err := db.Connect()
	assert.Nil(t, err)

	db.Drop()

	err = db.Initialize()

	go func() {
		time.Sleep(1 * time.Second)
		// Note Close instead of unlock
		db.Close()
	}()

	err = db.Lock(10000)
	assert.Nil(t, err)

	db2 := createDatabase()

	err = db2.Connect()
	assert.Nil(t, err)

	defer db2.Close()

	// The function below will block until db.Close() is called in the go-routine above
	err = db2.Lock(10000)
	assert.Nil(t, err)
}

func TestLockTimeout(t *testing.T) {
	db := createDatabase()

	err := db.Connect()
	assert.Nil(t, err)

	db.Drop()

	err = db.Initialize()

	go func() {
		time.Sleep(1 * time.Second)
		db.Close()
	}()

	err = db.Lock(10000)
	assert.Nil(t, err)

	db2 := createDatabase()

	err = db2.Connect()
	assert.Nil(t, err)

	defer db2.Close()

	err = db2.Lock(100)
	assert.NotNil(t, err) // We should get an locked request timed out error
}
