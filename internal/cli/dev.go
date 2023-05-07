package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/monitoring"
	"github.com/colonyos/colonies/pkg/server"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gin-gonic/gin"

	"github.com/colonyos/colonies/pkg/security"
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

		envErr := false

		if os.Getenv("LANG") == "" {
			log.Error("LANG environmental variable missing, try export LANG=en_US.UTF-8")
			envErr = true
		}

		if os.Getenv("LANGUAGE") == "" {
			log.Error("LANGUAGE environmental variable missing, try export LANGUAGE=en_US.UTF-8")
			envErr = true
		}

		if os.Getenv("LC_ALL") == "" {
			log.Error("LC_ALL environmental variable missing, try export LC_ALL=en_US.UTF-8")
			envErr = true
		}

		if os.Getenv("LC_CTYPE") == "" {
			log.Error("LC_CTYPE environmental variable missing, try export LC_CTYPE=UTF-8")
			envErr = true
		}

		if os.Getenv("TZ") == "" {
			log.Error("TZ environmental variable missing, try export TZ=Europe/Stockholm")
			envErr = true
		}

		if os.Getenv("COLONIES_SERVER_HOST") == "" {
			log.Error("COLONIES_SERVER_HOST environmental variable missing, try export COLONIES_SERVER_HOST=\"localhost\"")
			envErr = true
		}

		if os.Getenv("COLONIES_SERVER_PORT") == "" {
			log.Error("COLONIES_SERVER_PORT environmental variable missing, try export COLONIES_SERVER_PORT=\"50080\"")
			envErr = true
		}

		if os.Getenv("COLONIES_MONITOR_PORT") == "" {
			log.Error("COLONIES_MONITOR_PORT environmental variable missing, try export COLONIES_MONITOR_PORT=\"21120\"")
			envErr = true
		}

		if os.Getenv("COLONIES_MONITOR_INTERVAL") == "" {
			log.Error("COLONIES_MONITOR_INTERVAL environmental variable missing, try export COLONIES_MONITOR_INTERVAL=\"1\"")
			envErr = true
		}

		if os.Getenv("COLONIES_SERVER_ID") == "" {
			log.Error("COLONIES_SERVER_ID environmental variable missing, try export COLONIES_SERVER_ID=\"039231c7644e04b6895471dd5335cf332681c54e27f81fac54f9067b3f2c0103\"")
			envErr = true
		}

		if os.Getenv("COLONIES_SERVER_PRVKEY") == "" {
			log.Error("COLONIES_SERVER_PRVKEY environmental variable missing, try export COLONIES_SERVER_PRVKEY=\"fcc79953d8a751bf41db661592dc34d30004b1a651ffa0725b03ac227641499d\"")
			envErr = true
		}

		if os.Getenv("COLONIES_DB_HOST") == "" {
			log.Error("COLONIES_DB_HOST environmental variable missing, try export COLONIES_DB_HOST=\"localhost\"")
			envErr = true
		}

		if os.Getenv("COLONIES_DB_USER") == "" {
			log.Error("COLONIES_DB_USER environmental variable missing, try export COLONIES_DB_USER=\"postgres\"")
			envErr = true
		}

		if os.Getenv("COLONIES_DB_PORT") == "" {
			log.Error("COLONIES_DB_PORT environmental variable missing, try export COLONIES_DB_PORT=\"50070\"")
			envErr = true
		}

		if os.Getenv("COLONIES_DB_PASSWORD") == "" {
			log.Error("COLONIES_DB_PASSWORD environmental variable missing, try export COLONIES_DB_PASSWORD=\"rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7\"")
			envErr = true
		}

		if os.Getenv("COLONIES_COLONY_ID") == "" {
			log.Error("COLONIES_COLONY_ID environmental variable missing, try export COLONIES_COLONY_ID=\"4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4\"")
			envErr = true
		}

		if os.Getenv("COLONIES_COLONY_PRVKEY") == "" {
			log.Error("COLONIES_COLONY_PRVKEY environmental variable missing, try export COLONIES_COLONY_PRVKEY=\"ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514\"")
			envErr = true
		}

		if os.Getenv("COLONIES_EXECUTOR_ID") == "" {
			log.Error("COLONIES_EXECUTOR_ID environmental variable missing, try export COLONIES_EXECUTOR_ID=\"3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac\"")
			envErr = true
		}

		if os.Getenv("COLONIES_EXECUTOR_PRVKEY") == "" {
			log.Error("COLONIES_EXECUTOR_PRVKEY environmental variable missing, try export COLONIES_EXECUTOR_PRVKEY=\"ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05\"")
			envErr = true
		}

		if os.Getenv("COLONIES_EXECUTOR_TYPE") == "" {
			log.Error("COLONIES_EXECUTOR_TYPE environmental variable missing, try export COLONIES_EXECUTOR_TYPE=\"cli\"")
			envErr = true
		}

		if envErr {
			log.Error(envErr)
			fmt.Println("\nExample of enironmental variables:")
			envProposal := "export LANG=en_US.UTF-8\n"
			envProposal += "export LANGUAGE=en_US.UTF-8\n"
			envProposal += "export LC_ALL=en_US.UTF-8\n"
			envProposal += "export LC_CTYPE=UTF-8\n"
			envProposal += "export TZ=Europe/Stockholm\n"
			envProposal += "export COLONIES_TLS=\"false\"\n"
			envProposal += "export COLONIES_SERVER_HOST=\"localhost\"\n"
			envProposal += "export COLONIES_SERVER_PORT=\"50080\"\n"
			envProposal += "export COLONIES_MONITOR_PORT=\"21120\"\n"
			envProposal += "export COLONIES_SERVER_ID=\"039231c7644e04b6895471dd5335cf332681c54e27f81fac54f9067b3f2c0103\"\n"
			envProposal += "export COLONIES_SERVER_PRVKEY=\"fcc79953d8a751bf41db661592dc34d30004b1a651ffa0725b03ac227641499d\"\n"
			envProposal += "export COLONIES_DB_HOST=\"localhost\"\n"
			envProposal += "export COLONIES_DB_USER=\"postgres\"\n"
			envProposal += "export COLONIES_DB_PORT=\"50070\"\n"
			envProposal += "export COLONIES_DB_PASSWORD=\"rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7\"\n"
			envProposal += "export COLONIES_COLONY_ID=\"4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4\"\n"
			envProposal += "export COLONIES_COLONY_PRVKEY=\"ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514\"\n"
			envProposal += "export COLONIES_EXECUTOR_ID=\"3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac\"\n"
			envProposal += "export COLONIES_EXECUTOR_PRVKEY=\"ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05\"\n"
			envProposal += "export COLONIES_EXECUTOR_TYPE=\"cli\"\n"
			fmt.Println(envProposal)
			os.Exit(-1)
		}

		coloniesPath := "/tmp/coloniesdev/"
		log.WithFields(log.Fields{"Path": coloniesPath}).Info("Creating Colonies data directory, this directory will be deleted every time the development server is restarted")
		err := os.Mkdir(coloniesPath, 0700)
		if err != nil {
			os.RemoveAll(coloniesPath)
			err = os.Mkdir(coloniesPath, 0700)
			CheckError(err)
		}

		dbHost := os.Getenv("COLONIES_DB_HOST")
		dbPort, err := strconv.Atoi(os.Getenv("COLONIES_DB_PORT"))
		CheckError(err)

		dbUser := os.Getenv("COLONIES_DB_USER")
		dbPassword := os.Getenv("COLONIES_DB_PASSWORD")

		AllowExecutorReregisterStr := os.Getenv("COLONIES_ALLOW_EXECUTOR_REREGISTER")
		if AllowExecutorReregisterStr != "" {
			AllowExecutorReregister, err = strconv.ParseBool(AllowExecutorReregisterStr)
			CheckError(err)
		} else {
			AllowExecutorReregister = false
		}

		log.WithFields(log.Fields{"DBHost": dbHost, "DBPort": dbPort, "DBUser": dbUser, "DBPassword": dbPassword, "DBName": DBName}).Info("Starting embedded PostgreSQL server")
		postgres := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
			RuntimePath(coloniesPath + "/embedded-postgres-go/extracted").
			BinariesPath(coloniesPath + "/embedded-postgres-go/extracted").
			DataPath(coloniesPath + "/embedded-postgres-go/extracted/data").
			Username(dbUser).
			Version(embeddedpostgres.V14).
			Password(dbPassword).
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

		log.WithFields(log.Fields{"DBHost": dbHost, "DBPort": dbPort, "DBUser": dbUser, "DBPassword": dbPassword, "DBName": DBName}).Info("Connecting to PostgreSQL server")
		coloniesDB := postgresql.CreatePQDatabase(dbHost, dbPort, dbUser, dbPassword, DBName, DBPrefix)
		err = coloniesDB.Connect()
		CheckError(err)

		log.Info("Initialize a Colonies PostgreSQL database")
		err = coloniesDB.Initialize()
		CheckError(err)

		keychain, err := security.CreateKeychain(".colonies")
		CheckError(err)

		serverID := os.Getenv("COLONIES_SERVER_ID")
		serverPrvKey := os.Getenv("COLONIES_SERVER_PRVKEY")
		log.WithFields(log.Fields{"ServerId": serverID, "ServerPrvKey": serverPrvKey}).Info("Adding a ServerId to keychain")
		err = keychain.AddPrvKey(serverID, serverPrvKey)
		CheckError(err)

		colonyID := os.Getenv("COLONIES_COLONY_ID")
		colonyPrvKey := os.Getenv("COLONIES_COLONY_PRVKEY")
		log.WithFields(log.Fields{"ColonyId": colonyID, "ColonyPrvKey": colonyPrvKey}).Info("Adding a ColonyId to keychain")
		err = keychain.AddPrvKey(colonyID, colonyPrvKey)
		CheckError(err)

		executorID := os.Getenv("COLONIES_EXECUTOR_ID")
		executorPrvKey := os.Getenv("COLONIES_EXECUTOR_PRVKEY")
		log.WithFields(log.Fields{"ExecutorId": executorID, "ExecutorPrvKey": executorPrvKey}).Info("Adding a ExecutorId to keychain")
		err = keychain.AddPrvKey(executorID, executorPrvKey)
		CheckError(err)

		coloniesServerPort, err := strconv.Atoi(os.Getenv("COLONIES_SERVER_PORT"))
		CheckError(err)
		log.WithFields(log.Fields{"Port": coloniesServerPort}).Info("Starting a Colonies server")

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

		ExclusiveAssignEnvStr := os.Getenv("COLONIES_EXCLUSIVE_ASSIGN")
		if ExclusiveAssignEnvStr != "" {
			ExclusiveAssign, err = strconv.ParseBool(ExclusiveAssignEnvStr)
			CheckError(err)
		} else {
			ExclusiveAssign = false
		}

		node := cluster.Node{Name: "dev", Host: "localhost", APIPort: coloniesServerPort, EtcdClientPort: 2379, EtcdPeerPort: 2380, RelayPort: 2381}
		clusterConfig := cluster.Config{}
		clusterConfig.AddNode(node)

		if Verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			gin.SetMode(gin.ReleaseMode)
			gin.DefaultWriter = ioutil.Discard
		}

		retentionStr := os.Getenv("COLONIES_RETENTION")
		retention := false
		if retentionStr == "true" {
			retention = true
		}
		retentionPolicyStr := os.Getenv("COLONIES_RETENTION_POLICY")
		retentionPolicy, err := strconv.ParseInt(retentionPolicyStr, 10, 64)
		CheckError(err)

		retentionPeriod := 60000 // Run retention worker once a minute

		coloniesServer := server.CreateColoniesServer(coloniesDB,
			coloniesServerPort,
			serverID,
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
			retention,
			retentionPolicy,
			retentionPeriod)

		go coloniesServer.ServeForever()

		coloniesServerHost := os.Getenv("COLONIES_SERVER_HOST")
		log.WithFields(log.Fields{"ColoniesServerHost": coloniesServerHost, "ColoniesServerPort": coloniesServerPort}).Info("Connecting to Colonies server")
		client := client.CreateColoniesClient(coloniesServerHost, coloniesServerPort, true, false)

		log.WithFields(log.Fields{"ColonyID": colonyID}).Info("Registering a new Colony")
		colony := core.CreateColony(colonyID, "dev")
		_, err = client.AddColony(colony, serverPrvKey)
		CheckError(err)

		executorType := os.Getenv("COLONIES_EXECUTOR_TYPE")
		executorName := "myexecutor"
		log.WithFields(log.Fields{"ExecutorID": executorID, "ExecutorType": executorType, "ExecutorName": executorName}).Info("Registering a new executor")

		executor := core.CreateExecutor(executorID, executorType, executorName, colonyID, time.Now(), time.Now())
		_, err = client.AddExecutor(executor, colonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ExecutorID": executorID}).Info("Approving executor")
		log.Info("Approving CLI executor")
		err = client.ApproveExecutor(executorID, colonyPrvKey)
		CheckError(err)

		monitorPortStr := os.Getenv("COLONIES_MONITOR_PORT")
		monitorPort, err := strconv.Atoi(monitorPortStr)
		CheckError(err)

		intervalStr := os.Getenv("COLONIES_MONITOR_INTERVAL")
		interval, err := strconv.Atoi(intervalStr)
		CheckError(err)

		log.WithFields(log.Fields{
			"Port":               monitorPort,
			"ColoniesServerHost": coloniesServerHost,
			"ColoniesServerPort": coloniesServerPort,
			"PullInterval":       interval}).
			Info("Starting Prometheus monitoring server")
		monitoring.CreateMonitoringServer(monitorPort, coloniesServerHost, coloniesServerPort, true, true, serverPrvKey, interval)

		wait := make(chan bool)

		log.Info("Successfully started Colonies development server")
		log.Info("Press ctrl+c to exit")
		<-wait
	},
}
