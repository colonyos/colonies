package postgresql

import (
	"io/ioutil"
	"log"
)

func PrepareTests() (*PQDatabase, error) {
	log.SetOutput(ioutil.Discard)

	dbHost := "localhost"
	dbPort := 5432
	dbUser := "postgres"
	dbPassword := "rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7"
	dbName := "postgres"
	dbPrefix := "TEST_"

	db := CreatePQDatabase(dbHost, dbPort, dbUser, dbPassword, dbName, dbPrefix)

	err := db.Connect()
	if err != nil {
		return nil, err
	}

	err = db.Drop()
	if err != nil {
		// ignore
	}

	err = db.Initialize()

	return db, err
}
