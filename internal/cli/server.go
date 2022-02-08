package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/colonyos/colonies/internal/logging"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/spf13/cobra"
)

func init() {
	serverCmd.AddCommand(serverStartCmd)
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringVarP(&DBHost, "dbhost", "", "", "Colonies database host")
	serverCmd.PersistentFlags().IntVarP(&DBPort, "dbport", "", DefaultDBPort, "Colonies database port")
	serverCmd.PersistentFlags().StringVarP(&DBUser, "dbuser", "", "", "Colonies database user")
	serverCmd.PersistentFlags().StringVarP(&DBPassword, "dbpassword", "", "", "Colonies database password")
	serverCmd.PersistentFlags().StringVarP(&TLSCert, "tlscert", "", "", "TLS certificate")
	serverCmd.PersistentFlags().StringVarP(&TLSKey, "tlskey", "", "", "TLS key")
	serverCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", DefaultServerPort, "Server HTTP port")
	serverCmd.PersistentFlags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage a Colonies server",
	Long:  "Manage a Colonies server",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Colonies server",
	Long:  "Start a Colonies server",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if DBHost == "" {
			DBHost = os.Getenv("DBHOST")
		}

		if DBPort != DefaultDBPort {
			DBPortStr := os.Getenv("DBPORT")
			DBPort, err = strconv.Atoi(DBPortStr)
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

		if ServerPort != DefaultServerPort {
			ServerPortStr := os.Getenv("SERVERPORT")
			ServerPort, err = strconv.Atoi(ServerPortStr)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
		}

		if ServerID == "" {
			ServerID = os.Getenv("SERVERID")
		}

		if TLSKey == "" {
			TLSKey = os.Getenv("TLSKEY")
		}

		if TLSCert == "" {
			TLSCert = os.Getenv("TLSCERT")
		}

		VerboseEnv := os.Getenv("VERBOSE")
		if VerboseEnv == "true" {
			Verbose = true
		} else if VerboseEnv == "false" {
			Verbose = false
		}

		fmt.Println("----------")
		fmt.Println("DBHOST: ", DBHost)
		fmt.Println("DBPORT: ", DBPort)
		fmt.Println("DBUser: ", DBUser)
		fmt.Println("DBPassword: ", DBPassword)
		fmt.Println("DBName:", DBName)
		fmt.Println("DBPrefix:", DBPrefix)
		fmt.Println("----------")

		db := postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix)
		err = db.Connect()
		if err != nil {
			fmt.Println("Failed to connect to database")
			os.Exit(-1)
		}
		logging.Log().Info("Connecting to Colonies database, host: " + DBHost + ", port: " + strconv.Itoa(DBPort) + ", user: " + DBUser + ", password: " + "******************, name: " + DBName + ". prefix: " + DBPrefix)
		server := server.CreateColoniesServer(db, ServerPort, ServerID, TLSKey, TLSCert, Verbose)
		server.ServeForever()
	},
}
