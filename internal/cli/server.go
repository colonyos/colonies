package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/gin-gonic/gin"
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
	serverCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")
	serverCmd.PersistentFlags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	serverCmd.PersistentFlags().StringVarP(&EtcdName, "etcdname", "", "etcd", "Etcd name")
	serverCmd.PersistentFlags().StringVarP(&EtcdHost, "etcdhost", "", "0.0.0.0", "Etcd host name")
	serverCmd.PersistentFlags().IntVarP(&EtcdClientPort, "etcdclientport", "", 2379, "Etcd port")
	serverCmd.PersistentFlags().IntVarP(&EtcdPeerPort, "etcdpeerport", "", 2380, "Etcd peer port")
	serverCmd.PersistentFlags().IntVarP(&RelayPort, "relayport", "", 2381, "Colonies server relay port")
	serverCmd.PersistentFlags().StringSliceVarP(&EtcdCluster, "initial-cluster", "", make([]string, 0), "Cluster config, e.g. --etcdcluster server1=localhost:peerport:relayport:apiport,server2=localhost:peerport:relayport:apiport")
	serverCmd.PersistentFlags().StringVarP(&EtcdDataDir, "etcddatadir", "", "", "Etcd data dir")

	serverStatusCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	serverStatusCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

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
		if ServerPort == -1 {
			ServerPort, err = strconv.Atoi(ServerPortEnvStr)
			CheckError(err)
		}
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

	CronPeriodCheckerEnvStr := os.Getenv("COLONIES_CRON_CHECKER_PERIOD")
	if CronPeriodCheckerEnvStr != "" {
		CronCheckerPeriod, err = strconv.Atoi(CronPeriodCheckerEnvStr)
		CheckError(err)
	} else {
		CronCheckerPeriod = server.CRON_TRIGGER_PERIOD
	}

	GeneratorPeriodCheckerEnvStr := os.Getenv("COLONIES_GENERATOR_CHECKER_PERIOD")
	if GeneratorPeriodCheckerEnvStr != "" {
		GeneratorCheckerPeriod, err = strconv.Atoi(GeneratorPeriodCheckerEnvStr)
		CheckError(err)
	} else {
		GeneratorCheckerPeriod = server.GENERATOR_TRIGGER_PERIOD
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
		parseDBEnv()
		parseServerEnv()

		if !Insecure {
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

		log.WithFields(log.Fields{"DBHost": DBHost, "DBPort": DBPort, "DBUser": DBUser, "DBPassword": "*******************", "DBName": DBName, "UseTLS": UseTLS}).Info("Connecting to PostgreSQL database")
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

		node := cluster.Node{Name: EtcdName, Host: EtcdHost, APIPort: ServerPort, EtcdClientPort: EtcdClientPort, EtcdPeerPort: EtcdPeerPort, RelayPort: RelayPort}
		clusterConfig := cluster.Config{}

		if len(EtcdCluster) > 0 {
			// Parse EtcdCluster flag
			errMsg := "Invalid cluster, try e.g. --etcdcluster server1=localhost:23100:25100:26100,server2=localhost:23101:25101:26101"
			for _, s := range EtcdCluster {
				split1 := strings.Split(s, "=")
				if len(split1) != 2 {
					CheckError(errors.New(errMsg))
				}
				name := split1[0]
				split2 := strings.Split(split1[1], ":")
				if len(split2) != 4 {
					CheckError(errors.New(errMsg))
				}
				host := split2[0]
				portStr1 := split2[1]
				etcPeerPort, err := strconv.Atoi(portStr1)
				CheckError(err)
				portStr2 := split2[2]
				relayPort, err := strconv.Atoi(portStr2)
				CheckError(err)
				portStr3 := split2[3]
				apiPort, err := strconv.Atoi(portStr3)
				CheckError(err)
				node := cluster.Node{Name: name, Host: host, EtcdClientPort: EtcdClientPort, EtcdPeerPort: etcPeerPort, RelayPort: relayPort, APIPort: apiPort}
				clusterConfig.AddNode(node)
			}
		} else {
			clusterConfig.AddNode(node)
		}

		if EtcdDataDir == "" {
			EtcdDataDir = "/tmp/colonies/prod/etcd"
			log.Warning("EtcdDataDir not specified, setting it to " + EtcdDataDir)
		}

		log.WithFields(log.Fields{
			"BuildVersion": build.BuildVersion,
			"BuildTime":    build.BuildTime,
			"ServerPort":   ServerPort,
			"Verbose":      Verbose,
			"UseTLS":       UseTLS,
			"ServerID":     ServerID,
		}).Info("Starting a Colonies Server")

		if Verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			gin.SetMode(gin.ReleaseMode)
			gin.DefaultWriter = ioutil.Discard
		}

		server := server.CreateColoniesServer(db,
			ServerPort,
			ServerID,
			UseTLS,
			TLSKey,
			TLSCert,
			node,
			clusterConfig,
			EtcdDataDir,
			GeneratorCheckerPeriod,
			CronCheckerPeriod)

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
