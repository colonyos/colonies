package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/server"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"

	"github.com/colonyos/colonies/pkg/security"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(devCmd)
}

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start a Colonies development server",
	Long:  "Start a Colonies development server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting a Colonies development server")

		log.Info("Creating Colonies data dir")
		coloniesPath := "/tmp/colonies"
		err := os.Mkdir(coloniesPath, 0700)
		if err != nil {
			os.RemoveAll(coloniesPath)
			err = os.Mkdir(coloniesPath, 0700)
			CheckError(err)
		}

		dbHost := os.Getenv("COLONIES_DBHOST")
		dbPort, err := strconv.Atoi(os.Getenv("COLONIES_DBPORT"))
		CheckError(err)

		dbUser := os.Getenv("COLONIES_DBUSER")
		dbPassword := os.Getenv("COLONIES_DBPASSWORD")

		log.Info("Starting an embedded PostgreSQL server")
		postgres := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
			RuntimePath(coloniesPath + "/embedded-postgres-go/extracted").
			BinariesPath(coloniesPath + "/embedded-postgres-go/extracted").
			DataPath(coloniesPath + "/embedded-postgres-go/extracted/data").
			Username(dbUser).
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

		log.Info("Connecting to PostgreSQL server")
		coloniesDB := postgresql.CreatePQDatabase(dbHost, dbPort, dbUser, dbPassword, DBName, DBPrefix)
		err = coloniesDB.Connect()
		CheckError(err)

		log.Info("Initialize a Colonies PostgreSQL database")
		err = coloniesDB.Initialize()
		CheckError(err)

		keychain, err := security.CreateKeychain(".colonies")
		CheckError(err)

		log.Info("Adding a ServerId ")
		serverID := os.Getenv("COLONIES_SERVERID")
		serverPrvKey := os.Getenv("COLONIES_SERVERPRVKEY")
		err = keychain.AddPrvKey(serverID, serverPrvKey)
		CheckError(err)

		log.Info("Adding a ColonyId")
		colonyID := os.Getenv("COLONIES_COLONYID")
		colonyPrvKey := os.Getenv("COLONIES_COLONYPRVKEY")
		err = keychain.AddPrvKey(colonyID, colonyPrvKey)
		CheckError(err)

		log.Info("Adding a RuntimeId")
		runtimeID := os.Getenv("COLONIES_RUNTIMEID")
		runtimePrvKey := os.Getenv("COLONIES_RUNTIMEPRVKEY")
		err = keychain.AddPrvKey(runtimeID, runtimePrvKey)
		CheckError(err)

		coloniesServerPort, err := strconv.Atoi(os.Getenv("COLONIES_SERVERPORT"))
		CheckError(err)

		coloniesServer := server.CreateColoniesServer(coloniesDB, coloniesServerPort, serverID, false, "", "", true)
		go coloniesServer.ServeForever()

		client := client.CreateColoniesClient(os.Getenv("COLONIES_SERVERHOST"), coloniesServerPort, true, false)

		log.Info("Register a Colony")
		colony := core.CreateColony(colonyID, "dev")
		_, err = client.AddColony(colony, serverPrvKey)
		CheckError(err)

		log.Info("Registering a CLI runtime")

		runtimeType := os.Getenv("COLONIES_RUNTIMETYPE")
		runtimeGroup := os.Getenv("COLONIES_RUNTIMEGROUP")

		runtime := core.CreateRuntime(runtimeID, runtimeType, "dev_runtime", runtimeGroup, colonyID, "", 1, 0, "", 0, time.Now(), time.Now())
		_, err = client.AddRuntime(runtime, colonyPrvKey)
		CheckError(err)

		log.Info("Approving CLI runtime")
		err = client.ApproveRuntime(runtimeID, colonyPrvKey)
		CheckError(err)

		wait := make(chan bool)

		fmt.Println()

		log.Info("Add the following environmental variables to your shell:")
		envStr := "export LANG=en_US.UTF-8\n"
		envStr += "export LANGUAGE=en_US.UTF-8\n"
		envStr += "export LC_ALL=en_US.UTF-8\n"
		envStr += "export LC_CTYPE=UTF-8\n"
		envStr += "export TZ=Europe/Stockholm\n"
		envStr += "export COLONIES_SERVERPROTOCOL=\"http\"\n"
		envStr += "export COLONIES_SERVERHOST=\"localhost\"\n"
		envStr += "export COLONIES_SERVERPORT=\"50080\"\n"
		envStr += "export COLONIES_SERVERID=\"" + serverID + "\"\n"
		envStr += "export COLONIES_SERVERPRVKEY=\"" + serverPrvKey + "\"\n"
		envStr += "export COLONIES_DBHOST=\"localhost\"\n"
		envStr += "export COLONIES_DBUSER=\"postgres\"\n"
		envStr += "export COLONIES_DBPORT=\"50070\"\n"
		envStr += "expoer COLONIES_DBPASSWORD=\"rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7\"\n"
		envStr += "export COLONIES_COLONYID=\"" + colonyID + "\"\n"
		envStr += "export COLONIES_COLONYPRVKEY=\"" + colonyPrvKey + "\"\n"
		envStr += "export COLONIES_RUNTIMEID=\"" + runtimeID + "\"\n"
		envStr += "export COLONIES_RUNTIMEPRVKEY=\"" + runtimePrvKey + "\"\n"
		envStr += "export COLONIES_RUNTIMETYPE=\"" + runtimeType + "\"\n"
		envStr += "export COLONIES_RUNTIMEGROUP=\"" + runtimeGroup + "\"\n"

		fmt.Println(envStr)

		log.Info("Press ctrl+c to exit")
		<-wait
	},
}
