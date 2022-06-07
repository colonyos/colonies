package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStatusCmd)
	serverCmd.AddCommand(serverStatisticsCmd)
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

	serverStatisticsCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	serverStatisticsCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage production server",
	Long:  "Manage production server",
}

func parseServerEnv() {
	var err error
	ServerHostEnv := os.Getenv("COLONIES_SERVERHOST")
	if ServerHostEnv != "" {
		ServerHost = ServerHostEnv
	}

	ServerPortEnvStr := os.Getenv("COLONIES_SERVERPORT")
	if ServerPortEnvStr != "" {
		ServerPort, err = strconv.Atoi(ServerPortEnvStr)
		CheckError(err)
	}

	if ServerID == "" {
		ServerID = os.Getenv("COLONIES_SERVERID")
	}

	TLSEnv := os.Getenv("COLONIES_TLS")
	if TLSEnv == "true" {
		UseTLS = true
		Insecure = false
	} else if TLSEnv == "false" {
		UseTLS = false
		Insecure = true
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
	Short: "Show status about a production server",
	Long:  "Show status about a production server",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()
		parseServerEnv()

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

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
	Short: "Start a production server",
	Long:  "Start a production server",
	Run: func(cmd *cobra.Command, args []string) {
		log.WithFields(log.Fields{"BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Info("Starting a Colonies Server")
		parseDBEnv()
		parseServerEnv()

		fmt.Println(Insecure)

		if !Insecure {
			_, err := os.Stat(TLSKey)
			if err != nil {
				fmt.Println("XXXXXXX")
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

		log.WithFields(log.Fields{"DBHost": DBHost, "DBPort": DBPort, "DBUser": DBUser, "DBPassword": "*******************", "DBName": DBName, "UseTLS": UseTLS}).Info("Connecting to PostgreSQL database")
		server := server.CreateColoniesServer(db, ServerPort, ServerID, UseTLS, TLSKey, TLSCert, Verbose)
		for {
			err := server.ServeForever()
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Error("Failed to connect to Colonies Server")
				time.Sleep(1 * time.Second)
			}
		}
	},
}

var serverStatisticsCmd = &cobra.Command{
	Use:   "stat",
	Short: "Show server statistics",
	Long:  "Show server statistics",
	Run: func(cmd *cobra.Command, args []string) {
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

		stat, err := client.Statistics(ServerPrvKey)
		CheckError(err)

		fmt.Println("Process statistics:")
		specData := [][]string{
			[]string{"Colonies", strconv.Itoa(stat.Colonies)},
			[]string{"Runtimes", strconv.Itoa(stat.Runtimes)},
			[]string{"Waiting processes", strconv.Itoa(stat.WaitingProcesses)},
			[]string{"Running processes", strconv.Itoa(stat.RunningProcesses)},
			[]string{"Successful processes", strconv.Itoa(stat.SuccessfulProcesses)},
			[]string{"Failed processes", strconv.Itoa(stat.FailedProcesses)},
			[]string{"Waiting workflows", strconv.Itoa(stat.WaitingWorkflows)},
			[]string{"Running workflows ", strconv.Itoa(stat.RunningWorkflows)},
			[]string{"Successful workflows", strconv.Itoa(stat.SuccessfulWorkflows)},
			[]string{"Failed workflows", strconv.Itoa(stat.FailedWorkflows)},
		}
		specTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range specData {
			specTable.Append(v)
		}
		specTable.SetAlignment(tablewriter.ALIGN_LEFT)
		specTable.Render()
	},
}
