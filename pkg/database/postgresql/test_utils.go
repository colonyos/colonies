package postgresql

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
)

func PrepareTests() (*PQDatabase, error) {
	return PrepareTestsWithPrefix("TEST_")
}

func PrepareTestsWithPrefix(prefix string) (*PQDatabase, error) {
	log.SetOutput(ioutil.Discard)

	rand.Seed(time.Now().UTC().UnixNano())

	dbHost := os.Getenv("COLONIES_DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := 5432
	dbUser := os.Getenv("COLONIES_DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("COLONIES_DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7"
	}
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
