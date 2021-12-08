package database

import (
	"io/ioutil"
	"log"
)

func PrepareTests() (*Database, error) {
	log.SetOutput(ioutil.Discard)

	dbHost := "localhost"
	dbPort := 5432
	dbUser := "postgres"
	dbPassword := "rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7"
	dbName := "postgres"
	dbPrefix := "test"

	db := CreateDatabase(dbHost, dbPort, dbUser, dbPassword, dbName, dbPrefix)

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
