package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

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

func parseServerEnv() {
	var err error
	ServerHostEnv := os.Getenv("SERVERHOST")
	if ServerHostEnv != "" {
		ServerHost = ServerHostEnv
	}

	ServerPortEnvStr := os.Getenv("SERVERPORT")
	if ServerPortEnvStr != "" {
		ServerPort, err = strconv.Atoi(ServerPortEnvStr)
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
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Colonies server",
	Long:  "Start a Colonies server",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()
		parseServerEnv()

		_, err := os.Stat(TLSKey)
		if err != nil {
			fmt.Println("Failed to load TLS Key: " + TLSKey)
			os.Exit(-1)
		}

		_, err = os.Stat(TLSCert)
		if err != nil {
			fmt.Println("Failed to load TLS Cert: " + TLSCert)
			os.Exit(-1)
		}

		var db *postgresql.PQDatabase
		for {
			db = postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix)
			err := db.Connect()
			if err != nil {
				fmt.Println("Failed to connect to database")
				fmt.Println(err)
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		logging.Log().Info("Connecting to Colonies database, host: " + DBHost + ", port: " + strconv.Itoa(DBPort) + ", user: " + DBUser + ", password: " + "******************, name: " + DBName + ". prefix: " + DBPrefix)
		server := server.CreateColoniesServer(db, ServerPort, ServerID, TLSKey, TLSCert, Verbose)
		for {
			err := server.ServeForever()
			if err != nil {
				fmt.Println("Failed to start Colonies Server")
				fmt.Println(err)
				time.Sleep(1 * time.Second)
			}
		}
	},
}
