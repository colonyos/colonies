package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/database/postgresql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	dbCmd.AddCommand(dbCreateCmd)
	dbCmd.AddCommand(dbDropCmd)
	rootCmd.AddCommand(dbCmd)

	dbCmd.PersistentFlags().StringVarP(&DBHost, "dbhost", "", DefaultDBHost, "Colonies database host")
	dbCmd.PersistentFlags().IntVarP(&DBPort, "dbport", "", DefaultDBPort, "Colonies database port")
	dbCmd.PersistentFlags().StringVarP(&DBUser, "dbuser", "", "", "Colonies database user")
	dbCmd.PersistentFlags().StringVarP(&DBPassword, "dbpassword", "", "", "Colonies database password")
}

var dbCmd = &cobra.Command{
	Use:   "database",
	Short: "Manage internal database",
	Long:  "Manage internal database",
}

func parseDBEnv() {
	DBTypeEnv := os.Getenv("COLONIES_DB_TYPE")
	if DBTypeEnv != "" {
		DBType = DBTypeEnv
	}

	DBHostEnv := os.Getenv("COLONIES_DB_HOST")
	if DBHostEnv != "" {
		DBHost = DBHostEnv
	}

	var err error
	DBPortEnvStr := os.Getenv("COLONIES_DB_PORT")
	if DBPortEnvStr != "" {
		DBPort, err = strconv.Atoi(DBPortEnvStr)
		CheckError(err)
	}

	if DBUser == "" {
		DBUser = os.Getenv("COLONIES_DB_USER")
	}

	if DBPassword == "" {
		DBPassword = os.Getenv("COLONIES_DB_PASSWORD")
	}

	timescaleDBEnv := os.Getenv("COLONIES_DB_TIMESCALEDB")
	if timescaleDBEnv == "true" {
		TimescaleDB = true
	} else {
		TimescaleDB = false
	}

	initDBStr := os.Getenv("COLONIES_INITDB")
	if initDBStr == "true" {
		InitDB = true
	}
}

var dbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a database",
	Long:  "Create a database",
	Run: func(cmd *cobra.Command, args []string) {
		parseEnv()
		parseDBEnv()

		var db *postgresql.PQDatabase
		for {
			db = postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix, TimescaleDB)
			err := db.Connect()
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to call db.Connect(), retrying in 1 second ...")
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		log.WithFields(log.Fields{"Host": DBHost, "Port": DBPort, "User": DBUser, "Password": "**********************", "Prefix": DBPrefix}).Error("Connected to PostgreSQL database")

		err := db.Initialize()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("Failed to call db.Initialize()")
			os.Exit(0)
		}

		log.WithFields(log.Fields{"ServerID": ServerID}).Info("Setting server ID")
		CheckError(db.SetServerID("", ServerID))

		log.Info("Colonies database initialized")
	},
}

var dbDropCmd = &cobra.Command{
	Use:   "drop",
	Short: "Drop the database",
	Long:  "Drop the database",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()

		fmt.Print("WARNING!!! Are you sure you want to drop the database? This operation cannot be undone! (YES,no): ")

		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')

		if reply == "YES\n" {
			log.WithFields(log.Fields{"DBHost": DBHost, "DBPort": DBPort, "DBUser": DBUser, "DBPassword": "*******************", "DBName": DBName, "UseTLS": UseTLS, "TimescaleDB": TimescaleDB}).Info("Connecting to PostgreSQL database")

			db := postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix, TimescaleDB)
			err := db.Connect()
			CheckError(err)

			err = db.Drop()
			CheckError(err)
			log.Info("Colonies database dropped")
		} else {
			log.Info("Aborting ...")
		}
	},
}
