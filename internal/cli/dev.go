package cli

import (
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/colonyos/colonies/internal/crypto"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/monitoring"
	"github.com/colonyos/colonies/pkg/server"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(devCmd)
}

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start a development server",
	Long:  "Start a development server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting a Colonies development server")

		parseEnv()
		parseDBEnv()

		coloniesPath := "/tmp/coloniesdev/"
		log.WithFields(log.Fields{"Path": coloniesPath}).Info("Creating Colonies data directory, this directory will be deleted every time the development server is restarted")
		err := os.Mkdir(coloniesPath, 0700)
		if err != nil {
			os.RemoveAll(coloniesPath)
			err = os.Mkdir(coloniesPath, 0700)
			CheckError(err)
		}

		err = os.Mkdir(coloniesPath+"embedded-postgres-go", 0700)
		CheckError(err)
		err = os.Mkdir(coloniesPath+"embedded-postgres-go/extracted", 0700)
		CheckError(err)
		err = os.Mkdir(coloniesPath+"embedded-postgres-go/extracted/data", 0700)
		CheckError(err)

		log.WithFields(log.Fields{"DBHost": DBHost, "DBPort": DBPort, "DBUser": DBUser, "DBPassword": DBPassword, "DBName": DBName}).Info("Starting embedded PostgreSQL server")
		postgres := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
			RuntimePath(coloniesPath + "embedded-postgres-go/extracted").
			BinariesPath(coloniesPath + "embedded-postgres-go/extracted").
			DataPath(coloniesPath + "embedded-postgres-go/extracted/data").
			Username(DBUser).
			Version(embeddedpostgres.V12).
			Password(DBPassword).
			Port(50070))
		defer postgres.Stop()
		err = postgres.Start()
		CheckError(err)

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			postgres.Stop()
			log.Info("Colonies development server stopped")
			os.Exit(0)
		}()

		log.WithFields(log.Fields{"DBHost": DBHost, "DBPort": DBPort, "DBUser": DBUser, "DBPassword": DBPassword, "DBName": DBName}).Info("Connecting to PostgreSQL server")
		coloniesDB := postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix, false)
		err = coloniesDB.Connect()
		CheckError(err)

		log.Info("Initialize a Colonies PostgreSQL database")
		err = coloniesDB.Initialize()
		CheckError(err)

		log.WithFields(log.Fields{"Port": ServerPort}).Info("Starting a Colonies server")

		node := cluster.Node{Name: "dev", Host: "localhost", APIPort: ServerPort, EtcdClientPort: 23790, EtcdPeerPort: 23800, RelayPort: 2381}
		clusterConfig := cluster.Config{}
		clusterConfig.AddNode(node)

		if Verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			gin.SetMode(gin.ReleaseMode)
			gin.DefaultWriter = ioutil.Discard
		}

		retentionPeriod := 60000 // Run retention worker once a minute

		setupProfiler()

		serverIdentity, err := crypto.CreateIdendityFromString(ServerPrvKey)
		CheckError(err)

		coloniesServer := server.CreateColoniesServer(coloniesDB,
			ServerPort,
			serverIdentity.ID(),
			false,
			"",
			"",
			node,
			clusterConfig,
			"/tmp/coloniesdev/dev/etcd",
			GeneratorCheckerPeriod,
			CronCheckerPeriod,
			ExclusiveAssign,
			AllowExecutorReregister,
			Retention,
			RetentionPolicy,
			retentionPeriod)

		go coloniesServer.ServeForever()

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort}).Info("Connecting to Colonies server")
		client := client.CreateColoniesClient(ServerHost, ServerPort, true, false)

		log.WithFields(log.Fields{"ColonyID": ColonyID, "ServerPrvKey": ServerPrvKey}).Info("Registering a new Colony")
		colony := core.CreateColony(ColonyID, "dev")
		_, err = client.AddColony(colony, ServerPrvKey)
		CheckError(err)

		executorName := "myexecutor"

		identity, err := crypto.CreateIdendityFromString(PrvKey)
		CheckError(err)

		executorID := identity.ID()
		log.WithFields(log.Fields{"ExecutorID": executorID, "ExecutorType": ExecutorType, "ExecutorName": executorName, "ColonyPrvKey": ColonyPrvKey}).Info("Registering a new executor")

		executor := core.CreateExecutor(executorID, ExecutorType, executorName, ColonyID, time.Now(), time.Now())
		_, err = client.AddExecutor(executor, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ExecutorID": executorID}).Info("Approving executor")
		log.Info("Approving CLI executor")
		err = client.ApproveExecutor(executorID, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"Port":            MonitorPort,
			"ServerHost":      ServerHost,
			"ServerPort":      ServerPort,
			"MonitorInterval": MonitorInterval}).
			Info("Starting Prometheus monitoring server")
		monitoring.CreateMonitoringServer(MonitorPort, ServerHost, ServerPort, true, true, ServerPrvKey, MonitorInterval)

		wait := make(chan bool)

		log.Info("Successfully started Colonies development server")
		log.Info("Press ctrl+c to exit")
		<-wait
	},
}
