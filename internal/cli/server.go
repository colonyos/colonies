package cli

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/service"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	serverCmd.AddCommand(chServerIDCmd)
	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStatusCmd)
	serverCmd.AddCommand(serverStatisticsCmd)
	serverCmd.AddCommand(serverAliveCmd)
	rootCmd.AddCommand(serverCmd)

	chServerIDCmd.Flags().StringVarP(&TargetServerID, "serverid", "", "", "Server Id")
	chServerIDCmd.MarkFlagRequired("serverid")

	serverCmd.PersistentFlags().StringVarP(&DBHost, "dbhost", "", "", "Colonies database host")
	serverCmd.PersistentFlags().IntVarP(&DBPort, "dbport", "", DefaultDBPort, "Colonies database port")
	serverCmd.PersistentFlags().StringVarP(&DBUser, "dbuser", "", "", "Colonies database user")
	serverCmd.PersistentFlags().StringVarP(&DBPassword, "dbpassword", "", "", "Colonies database password")
	serverCmd.PersistentFlags().StringVarP(&TLSCert, "tlscert", "", "", "TLS certificate")
	serverCmd.PersistentFlags().StringVarP(&TLSKey, "tlskey", "", "", "TLS key")
	serverCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")
	serverCmd.PersistentFlags().StringVarP(&EtcdName, "etcdname", "", "etcd", "Etcd name")
	serverCmd.PersistentFlags().StringVarP(&EtcdHost, "etcdhost", "", "0.0.0.0", "Etcd host name")
	serverCmd.PersistentFlags().IntVarP(&EtcdClientPort, "etcdclientport", "", 2379, "Etcd port")
	serverCmd.PersistentFlags().IntVarP(&EtcdPeerPort, "etcdpeerport", "", 2380, "Etcd peer port")
	serverCmd.PersistentFlags().IntVarP(&RelayPort, "relayport", "", 2381, "Colonies server relay port")
	serverCmd.PersistentFlags().StringSliceVarP(&EtcdCluster, "initial-cluster", "", make([]string, 0), "Cluster config, e.g. --etcdcluster server1=localhost:peerport:relayport:apiport,server2=localhost:peerport:relayport:apiport")
	serverCmd.PersistentFlags().StringVarP(&EtcdDataDir, "etcddatadir", "", "", "Etcd data dir")
	serverCmd.PersistentFlags().BoolVarP(&InitDB, "initdb", "", false, "Initialize DB")

	serverStatusCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	serverStatusCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	serverStatisticsCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage production server",
	Long:  "Manage production server",
}

var chServerIDCmd = &cobra.Command{
	Use:   "chid",
	Short: "Change server Id",
	Long:  "Change server Id",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if len(TargetServerID) != 64 {
			CheckError(errors.New("Invalid server Id length"))
		}

		CheckError(client.ChangeServerID(TargetServerID, ServerPrvKey))

		log.WithFields(log.Fields{
			"ServerId": ServerID}).
			Info("Changed server Id")
	},
}

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status about a production server",
	Long:  "Show status about a production server",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()
		client := setup()

		serverBuildVersion, serverBuildTime, err := client.Version()
		CheckError(err)

		printServerStatusTable(serverBuildVersion, serverBuildTime)
	},
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a production server",
	Long:  "Start a production server",
	Run: func(cmd *cobra.Command, args []string) {
		parseDBEnv()
		parseEnv()

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

		var db *postgresql.PQDatabase
		for {
			db = postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix, TimescaleDB)
			err := db.Connect()
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to connect to PostgreSQL database")
				time.Sleep(1 * time.Second)
			} else {
				break
			}
		}

		log.WithFields(log.Fields{"DBHost": DBHost, "DBPort": DBPort, "DBUser": DBUser, "DBPassword": "*******************", "DBName": DBName, "UseTLS": UseTLS, "TimescaleDB": TimescaleDB}).Info("Connected to PostgreSQL database")

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

		retentionPeriod := 60000 // Run retention worker once a minute

		setupProfiler()

		server := service.CreateColoniesServer(db,
			ServerPort,
			UseTLS,
			TLSKey,
			TLSCert,
			node,
			clusterConfig,
			EtcdDataDir,
			GeneratorCheckerPeriod,
			CronCheckerPeriod,
			ExclusiveAssign,
			AllowExecutorReregister,
			Retention,
			RetentionPolicy,
			retentionPeriod)

		if InitDB {
			err := db.Initialize()
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to call db.Initialize()")
			} else {
				log.WithFields(log.Fields{"ServerID": ServerID}).Info("Setting server ID")
				CheckError(db.SetServerID("", ServerID))

				log.Info("Colonies database initialized")
			}
		}

		for {
			err := server.ServeForever()
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Error("Failed to start Colonies Server")
				time.Sleep(1 * time.Second)
			}
		}
	},
}

var serverStatisticsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show server statistics",
	Long:  "Show server statistics",
	Run: func(cmd *cobra.Command, args []string) {
		parseEnv()

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Debug("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		stat, err := client.Statistics(ServerPrvKey)
		CheckError(err)

		printServerStatTable(stat)
	},
}

var serverAliveCmd = &cobra.Command{
	Use:   "alive",
	Short: "Check if a server is alive",
	Long:  "Check if a server is alive",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(0)
	},
}
