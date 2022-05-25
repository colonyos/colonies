package cli

import (
	"fmt"
	"os"
	"os/signal"
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
		log.Info("Starting a Colonies development Server")

		log.Info("Creating BEMIS data dir")
		coloniesPath := "/tmp/colonies"
		err := os.Mkdir(coloniesPath, 0700)
		if err != nil {
			os.RemoveAll(coloniesPath)
			err = os.Mkdir(coloniesPath, 0700)
			CheckError(err)
		}

		log.Info("Starting an embedded PostgreSQL server")
		postgres := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
			RuntimePath(coloniesPath + "/embedded-postgres-go/extracted").
			BinariesPath(coloniesPath + "/embedded-postgres-go/extracted").
			DataPath(coloniesPath + "/embedded-postgres-go/extracted/data"))
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
		coloniesDB := postgresql.CreatePQDatabase("localhost", 5432, "postgres", "postgres", DBName, DBPrefix)
		err = coloniesDB.Connect()
		CheckError(err)

		log.Info("Initialize a Colonies PostgreSQL database")
		err = coloniesDB.Initialize()
		CheckError(err)

		keychain, err := security.CreateKeychain(".colonies")
		CheckError(err)

		log.Info("Adding a ServerId ")
		serverID := "039231c7644e04b6895471dd5335cf332681c54e27f81fac54f9067b3f2c0103"
		serverPrvKey := "fcc79953d8a751bf41db661592dc34d30004b1a651ffa0725b03ac227641499d"
		err = keychain.AddPrvKey(serverID, serverPrvKey)
		CheckError(err)

		log.Info("Adding a ColonyId")
		colonyID := "4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4"
		colonyPrvKey := "ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514"
		err = keychain.AddPrvKey(colonyID, colonyPrvKey)
		CheckError(err)

		log.Info("Adding a RuntimeId")
		runtimeID := "3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac"
		runtimePrvKey := "ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05"
		err = keychain.AddPrvKey(runtimeID, runtimePrvKey)
		CheckError(err)

		log.Info("Starting a Colonies server")
		coloniesServer := server.CreateColoniesServer(coloniesDB, 50080, serverID, false, "", "", true)
		go coloniesServer.ServeForever()

		log.Info("Starting a Colonies client")
		client := client.CreateColoniesClient("localhost", 50080, true, false)

		log.Info("Register a Colony")
		colony := core.CreateColony(colonyID, "dev")
		_, err = client.AddColony(colony, serverPrvKey)
		CheckError(err)

		log.Info("Registering a CLI runtime")
		runtime := core.CreateRuntime(runtimeID, "cli", "cli", colonyID, "", 1, 0, "", 0, time.Now(), time.Now())
		_, err = client.AddRuntime(runtime, colonyPrvKey)
		CheckError(err)

		log.Info("Approving CLI runtime")
		err = client.ApproveRuntime(runtimeID, colonyPrvKey)
		CheckError(err)

		wait := make(chan bool)

		fmt.Println()

		log.Info("Add the following environmental variables to your shell:")
		envStr := "export COLONIES_SERVER_PROTOCOL=\"http\"\n"
		envStr += "export COLONIES_SERVER_HOST=\"localhost\"\n"
		envStr += "export COLONIES_SERVER_PORT=\"50080\"\n"
		envStr += "export COLONIES_SERVERID=\"" + serverID + "\"\n"
		envStr += "export COLONIES_SERVERPRVKEY=\"" + serverPrvKey + "\"\n"
		envStr += "export COLONIES_COLONYID=\"" + colonyID + "\"\n"
		envStr += "export COLONIES_COLONYPRVKEY=\"" + colonyPrvKey + "\"\n"
		envStr += "export COLONIES_RUNTIMEID=\"" + runtimeID + "\"\n"
		envStr += "export COLONIES_RUNTIMEPRVKEY=\"" + runtimePrvKey + "\"\n"

		fmt.Println(envStr)

		log.Info("Press ctrl+c to exit")
		<-wait
	},
}
