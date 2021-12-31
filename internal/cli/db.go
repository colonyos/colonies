package cli

import (
	"bufio"
	"colonies/internal/logging"
	"colonies/pkg/database/postgresql"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	dbCmd.AddCommand(dbCreateCmd)
	dbCmd.AddCommand(dbDropCmd)
	rootCmd.AddCommand(dbCmd)

	dbCmd.PersistentFlags().StringVarP(&DBHost, "dbhost", "", "localhost", "Colonies database host")
	dbCmd.MarkPersistentFlagRequired("dbhost")
	dbCmd.PersistentFlags().IntVarP(&DBPort, "dbport", "", 5432, "Colonies database port")
	dbCmd.MarkPersistentFlagRequired("dbport")
	dbCmd.PersistentFlags().StringVarP(&DBUser, "dbuser", "", "", "Colonies database user")
	dbCmd.MarkPersistentFlagRequired("dbuser")
	dbCmd.PersistentFlags().StringVarP(&DBPassword, "dbpassword", "", "", "Colonies database password")
	dbCmd.MarkPersistentFlagRequired("dbpassword")
}

var dbCmd = &cobra.Command{
	Use:   "database",
	Short: "manage Colonies database",
	Long:  "manage Colonies database",
}

var dbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create a database",
	Long:  "create a database",
	Run: func(cmd *cobra.Command, args []string) {
		db := postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix)
		err := db.Connect()
		if err != nil {
			fmt.Println("Failed to connect to database")
			os.Exit(-1)
		}
		logging.Log().Info("Connecting to Colonies database, host: " + DBHost + ", port: " + strconv.Itoa(DBPort) + ", user: " + DBUser + ", password: " + "******************, name: " + DBName + ". prefix: " + DBPrefix)
		err = db.Initialize()
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
