package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/colonyos/colonies/internal/logging"
	"github.com/colonyos/colonies/pkg/database/postgresql"
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
	DBHostEnv := os.Getenv("DBHOST")
	if DBHostEnv != "" {
		DBHost = DBHostEnv
	}

	var err error
	DBPortEnvStr := os.Getenv("DBPORT")
	if DBPortEnvStr != "" {
		DBPort, err = strconv.Atoi(DBPortEnvStr)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	if DBUser == "" {
		DBUser = os.Getenv("DBUSER")
	}

	if DBPassword == "" {
		DBPassword = os.Getenv("DBPASSWORD")
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
				fmt.Println("Failed to connect to database")
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}
		logging.Log().Info("Connecting to Colonies database, host: " + DBHost + ", port: " + strconv.Itoa(DBPort) + ", user: " + DBUser + ", password: " + "******************, name: " + DBName + ". prefix: " + DBPrefix)
		err := db.Initialize()
		if err != nil {
			fmt.Println("Failed to create database")
			os.Exit(-1)
		}
		logging.Log().Info("Colonies database created")
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
			if err != nil {
				fmt.Println("Failed to connect to database")
				os.Exit(-1)
			}
			logging.Log().Info("Connecting to Colonies database, host: " + DBHost + ", port: " + strconv.Itoa(DBPort) + ", user: " + DBUser + ", password: " + "******************, name: " + DBName + ". prefix: " + DBPrefix)
			err = db.Drop()
			if err != nil {
				fmt.Println("Failed to drop database")
				os.Exit(-1)
			}
			logging.Log().Info("Colonies database dropped")
		} else {
			fmt.Println("Aborting ...")
		}
	},
}
