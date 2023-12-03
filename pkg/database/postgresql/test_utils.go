package postgresql

import (
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

func PrepareTests() (*PQDatabase, error) {
	return PrepareTestsWithPrefix("TEST_")
}

func PrepareTestsWithPrefix(prefix string) (*PQDatabase, error) {
	log.SetOutput(ioutil.Discard)

	rand.Seed(time.Now().UTC().UnixNano())

	dbHost := "localhost"
	dbPort := 5432
	dbUser := "postgres"
	dbPassword := "rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7"
	dbName := "postgres"
	dbPrefix := prefix

	db := CreatePQDatabase(dbHost, dbPort, dbUser, dbPassword, dbName, dbPrefix, false)

	err := db.Connect()
	if err != nil {
		return nil, err
	}

	db.Drop()
	err = db.Initialize()

	return db, err
}
