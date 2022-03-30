package cli

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStatusCmd)
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringVarP(&DBHost, "dbhost", "", "", "Colonies database host")
	serverCmd.PersistentFlags().IntVarP(&DBPort, "dbport", "", DefaultDBPort, "Colonies database port")
	serverCmd.PersistentFlags().StringVarP(&DBUser, "dbuser", "", "", "Colonies database user")
	serverCmd.PersistentFlags().StringVarP(&DBPassword, "dbpassword", "", "", "Colonies database password")
	serverCmd.PersistentFlags().StringVarP(&TLSCert, "tlscert", "", "", "TLS certificate")
	serverCmd.PersistentFlags().StringVarP(&TLSKey, "tlskey", "", "", "TLS key")
	serverCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", DefaultServerPort, "Server HTTP port")
	serverCmd.PersistentFlags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")

	serverStatusCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	serverStatusCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 50080, "Server HTTP port")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage a Colonies Server",
	Long:  "Manage a Colonies Server",
}

func parseServerEnv() {
	var err error
	ServerHostEnv := os.Getenv("COLONIES_SERVER_HOST")
	if ServerHostEnv != "" {
		ServerHost = ServerHostEnv
	}

	ServerPortEnvStr := os.Getenv("COLONIES_SERVER_PORT")
	if ServerPortEnvStr != "" {
		ServerPort, err = strconv.Atoi(ServerPortEnvStr)
		CheckError(err)
	}

	if ServerID == "" {
		ServerID = os.Getenv("COLONIES_SERVERID")
	}

	TLSEnv := os.Getenv("COLONIES_TLS")
	if TLSEnv == "true" {
		TLS = true
	} else if TLSEnv == "false" {
		TLS = false
	}

	if TLSKey == "" {
		TLSKey = os.Getenv("COLONIES_TLSKEY")
	}

	if TLSCert == "" {
		TLSCert = os.Getenv("COLONIES_TLSCERT")
	}

	VerboseEnv := os.Getenv("COLONIES_VERBOSE")
	if VerboseEnv == "true" {
		Verbose = true
	} else if VerboseEnv == "false" {
		Verbose = false
	}
}

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status about a Colonies Server",
	Long:  "Show status about a Colonies Server",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()
		parseServerEnv()

		client := client.CreateColoniesClient(ServerHost, ServerPort, TLS, true) // XXX: Insecure

		serverBuildVersion, serverBuildTime, err := client.Version()
		CheckError(err)

		serverData := [][]string{
			[]string{"Server Host", ServerHost},
			[]string{"Server Port", strconv.Itoa(ServerPort)},
			[]string{"CLI Version", build.BuildVersion},
			[]string{"CLI BuildTime", build.BuildTime},
			[]string{"Server Version", serverBuildVersion},
			[]string{"Server BuildTime", serverBuildTime},
		}
		serverTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range serverData {
			serverTable.Append(v)
		}
		serverTable.SetAlignment(tablewriter.ALIGN_LEFT)
		serverTable.Render()
	},
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Colonies Server",
	Long:  "Start a Colonies Server",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Info("Starting a Colonies Server")
		parseDBEnv()
		parseServerEnv()

		if TLS {
			_, err := os.Stat(TLSKey)
			if err != nil {
				CheckError(errors.New("Failed to load TLS Key: " + TLSKey))
				os.Exit(-1)
			}

			_, err = os.Stat(TLSCert)
			if err != nil {
				CheckError(errors.New("Failed to load TLS Cert: " + TLSCert))
				os.Exit(-1)
			}
		}

		var db *postgresql.PQDatabase
		for {
			db = postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix)
			err := db.Connect()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Failed to connect to PostgreSQL database")
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		log.WithFields(log.Fields{"DBHost": DBHost, "DBPort": DBPort, "DBUser": DBUser, "DBPassword": "*******************", "DBName": DBName, "TLS": TLS}).Info("Connecting to PostgreSQL database")
		server := server.CreateColoniesServer(db, ServerPort, ServerID, TLS, TLSKey, TLSCert, Verbose)
		for {
			err := server.ServeForever()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Failed to connect to Colonies Server")
				time.Sleep(1 * time.Second)
			}
		}
	},
}
