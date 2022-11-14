package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/security"
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
	Short: "Manage PostgreSQL database",
	Long:  "Manage PostgreSQL database",
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
	Short: "Create a database",
	Long:  "Create a database",
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
	Short: "Drop the database",
	Long:  "Drop the database",
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

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Remotely reset the database",
	Long:  "Remotely reset the database",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ServerID == "" {
			ServerID = os.Getenv("COLONIES_SERVERID")
		}
		if ServerID == "" {
			CheckError(errors.New("Unknown Server Id"))
		}

		if ServerPrvKey == "" {
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

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
