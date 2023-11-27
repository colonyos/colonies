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
	dbCmd.AddCommand(dbResetCmd)
	rootCmd.AddCommand(dbCmd)

	dbCmd.PersistentFlags().StringVarP(&DBHost, "dbhost", "", DefaultDBHost, "Colonies database host")
	dbCmd.PersistentFlags().IntVarP(&DBPort, "dbport", "", DefaultDBPort, "Colonies database port")
	dbCmd.PersistentFlags().StringVarP(&DBUser, "dbuser", "", "", "Colonies database user")
	dbCmd.PersistentFlags().StringVarP(&DBPassword, "dbpassword", "", "", "Colonies database password")

	dbResetCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	dbResetCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
}

var dbCmd = &cobra.Command{
	Use:   "database",
	Short: "Manage internal database",
	Long:  "Manage internal database",
}

func parseDBEnv() {
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
}

var dbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a database",
	Long:  "Create a database",
	Run: func(cmd *cobra.Command, args []string) {
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

		log.WithFields(log.Fields{"Host": DBHost, "Port": DBPort, "User": DBUser, "Password": "**********************", "Prefix": DBPrefix}).Error("Connecting to PostgreSQL database")
		err := db.Initialize()
		if err != nil {
			log.Warning("Failed to create database")
			os.Exit(0)
		}
		log.Info("Colonies database created")
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

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Remotely reset the database",
	Long:  "Remotely reset the database",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()
		client := setup()

		fmt.Print("WARNING!!! Are you sure you want to reset the database? This operation cannot be undone! (YES,no): ")

		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')

		if reply == "YES\n" {
			client.ResetDatabase(ServerPrvKey)
			log.Info("Colonies database reset")
		} else {
			log.Info("Aborting ...")
		}
	},
}
