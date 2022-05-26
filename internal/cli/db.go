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
	Short: "manage Colonies database",
	Long:  "manage Colonies database",
}

func parseDBEnv() {
	DBHostEnv := os.Getenv("COLONIES_DBHOST")
	if DBHostEnv != "" {
		DBHost = DBHostEnv
	}

	var err error
	DBPortEnvStr := os.Getenv("COLONIES_DBPORT")
	if DBPortEnvStr != "" {
		DBPort, err = strconv.Atoi(DBPortEnvStr)
		CheckError(err)
	}

	if DBUser == "" {
		DBUser = os.Getenv("COLONIES_DBUSER")
	}

	if DBPassword == "" {
		DBPassword = os.Getenv("COLONIES_DBPASSWORD")
	}
}

var dbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a database",
	Long:  "create a database",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()

		var db *postgresql.PQDatabase
		for {
			db = postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix)
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
	Short: "drop the database",
	Long:  "drop the database",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()

		fmt.Print("WARNING!!! Are you sure you want to drop the database? This operation cannot be undone! (YES,no): ")

		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')

		if reply == "YES\n" {
			db := postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix)
			err := db.Connect()
			CheckError(err)
			log.Info("Connecting to Colonies database, host: " + DBHost + ", port: " + strconv.Itoa(DBPort) + ", user: " + DBUser + ", password: " + "******************, name: " + DBName + ". prefix: " + DBPrefix)
			err = db.Drop()
			CheckError(err)
			log.Info("Colonies database dropped")
		} else {
			log.Info("Aborting ...")
		}
	},
}
