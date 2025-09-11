package database

import (
	"fmt"

	"github.com/colonyos/colonies/pkg/database/postgresql"
	log "github.com/sirupsen/logrus"
)

type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgresql"
)

type DatabaseConfig struct {
	Type        DatabaseType
	Host        string
	Port        int
	User        string
	Password    string
	Name        string
	Prefix      string
	TimescaleDB bool

	DataDir string // Future use
}

func CreateDatabase(config DatabaseConfig) (Database, error) {
	log.WithFields(log.Fields{
		"DatabaseType": config.Type,
		"Host":         config.Host,
		"Port":         config.Port,
		"Name":         config.Name,
		"Prefix":       config.Prefix,
		"TimescaleDB":  config.TimescaleDB,
		"DataDir":      config.DataDir,
	}).Info("Creating database connection")

	switch config.Type {
	case PostgreSQL:
		log.WithFields(log.Fields{
			"Host":        config.Host,
			"Port":        config.Port,
			"Name":        config.Name,
			"TimescaleDB": config.TimescaleDB,
		}).Info("Initializing PostgreSQL database")

		db := postgresql.CreatePQDatabase(config.Host, config.Port, config.User, config.Password, config.Name, config.Prefix, config.TimescaleDB)
		return db, nil

	default:
		log.WithField("DatabaseType", config.Type).Error("Unsupported database type requested")
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
