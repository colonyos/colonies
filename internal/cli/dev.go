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

		if os.Getenv("COLONIES_SERVERPROTOCOL") == "" {
			log.Error("COLONIES_SERVERPROTOCOL environmental variable missing, try export COLONIES_SERVERPROTOCOL=\"http\"")
			envErr = true
		}

		if os.Getenv("COLONIES_SERVERHOST") == "" {
			log.Error("COLONIES_SERVERHOST environmental variable missing, try export COLONIES_SERVERHOST=\"localhost\"")
			envErr = true
		}

		if os.Getenv("COLONIES_SERVERPORT") == "" {
			log.Error("COLONIES_SERVERPORT environmental variable missing, try export COLONIES_SERVERPORT=\"50080\"")
			envErr = true
		}

		if os.Getenv("COLONIES_SERVERID") == "" {
			log.Error("COLONIES_SERVERID environmental variable missing, try export COLONIES_SERVERID=\"039231c7644e04b6895471dd5335cf332681c54e27f81fac54f9067b3f2c0103\"")
			envErr = true
		}

		if os.Getenv("COLONIES_SERVERPRVKEY") == "" {
			log.Error("COLONIES_SERVERPRVKEY environmental variable missing, try export COLONIES_SERVERPRVKEY=\"fcc79953d8a751bf41db661592dc34d30004b1a651ffa0725b03ac227641499d\"")
			envErr = true
		}

		if os.Getenv("COLONIES_DBHOST") == "" {
			log.Error("COLONIES_DBHOST environmental variable missing, try export COLONIES_DBHOST=\"localhost\"")
			envErr = true
		}

		if os.Getenv("COLONIES_DBUSER") == "" {
			log.Error("COLONIES_DBUSER environmental variable missing, try export COLONIES_DBUSER=\"postgres\"")
			envErr = true
		}

		if os.Getenv("COLONIES_DBPORT") == "" {
			log.Error("COLONIES_DBPORT environmental variable missing, try export COLONIES_DBPORT=\"50070\"")
			envErr = true
		}

		if os.Getenv("COLONIES_DBPASSWORD") == "" {
			log.Error("COLONIES_DBPASSWORD environmental variable missing, try export COLONIES_DBPASSWORD=\"rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7\"")
			envErr = true
		}

		if os.Getenv("COLONIES_COLONYID") == "" {
			log.Error("COLONIES_COLONYID environmental variable missing, try export COLONIES_COLONYID=\"4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4\"")
			envErr = true
		}

		if os.Getenv("COLONIES_COLONYPRVKEY") == "" {
			log.Error("COLONIES_COLONYPRVKEY environmental variable missing, try export COLONIES_COLONYPRVKEY=\"ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514\"")
			envErr = true
		}

		if os.Getenv("COLONIES_RUNTIMEID") == "" {
			log.Error("COLONIES_RUNTIMEID environmental variable missing, try export COLONIES_RUNTIMEID=\"3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac\"")
			envErr = true
		}

		if os.Getenv("COLONIES_RUNTIMEPRVKEY") == "" {
			log.Error("COLONIES_RUNTIMEPRVKEY environmental variable missing, try export COLONIES_RUNTIMEPRVKEY=\"ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05\"")
			envErr = true
		}

		if os.Getenv("COLONIES_RUNTIMETYPE") == "" {
			log.Error("COLONIES_RUNTIMETYPE environmental variable missing, try export COLONIES_RUNTIMETYPE=\"cli\"")
			envErr = true
		}

		if os.Getenv("COLONIES_RUNTIMEGROUP") == "" {
			log.Error("COLONIES_RUNTIMEGROUP environmental variable missing, try export COLONIES_RUNTIMEGROUP=\"local\"")
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
			envProposal += "export COLONIES_SERVERPROTOCOL=\"http\"\n"
			envProposal += "export COLONIES_SERVERHOST=\"localhost\"\n"
			envProposal += "export COLONIES_SERVERPORT=\"50080\"\n"
			envProposal += "export COLONIES_SERVERID=\"039231c7644e04b6895471dd5335cf332681c54e27f81fac54f9067b3f2c0103\"\n"
			envProposal += "export COLONIES_SERVERPRVKEY=\"fcc79953d8a751bf41db661592dc34d30004b1a651ffa0725b03ac227641499d\"\n"
			envProposal += "export COLONIES_DBHOST=\"localhost\"\n"
			envProposal += "export COLONIES_DBUSER=\"postgres\"\n"
			envProposal += "export COLONIES_DBPORT=\"50070\"\n"
			envProposal += "expoer COLONIES_DBPASSWORD=\"rFcLGNkgsNtksg6Pgtn9CumL4xXBQ7\"\n"
			envProposal += "export COLONIES_COLONYID=\"4787a5071856a4acf702b2ffcea422e3237a679c681314113d86139461290cf4\"\n"
			envProposal += "export COLONIES_COLONYPRVKEY=\"ba949fa134981372d6da62b6a56f336ab4d843b22c02a4257dcf7d0d73097514\"\n"
			envProposal += "export COLONIES_RUNTIMEID=\"3fc05cf3df4b494e95d6a3d297a34f19938f7daa7422ab0d4f794454133341ac\"\n"
			envProposal += "export COLONIES_RUNTIMEPRVKEY=\"ddf7f7791208083b6a9ed975a72684f6406a269cfa36f1b1c32045c0a71fff05\"\n"
			envProposal += "export COLONIES_RUNTIMETYPE=\"cli\"\n"
			envProposal += "export COLONIES_RUNTIMEGROUP=\"local\"\n"
			fmt.Println(envProposal)
			os.Exit(-1)
		}

		coloniesPath := "/tmp/colonies"
		log.WithFields(log.Fields{"Path": coloniesPath}).Info("Creating Colonies data directory, this directory will be deleted every time the development server is restarted")
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

		log.WithFields(log.Fields{"DBHost": dbHost, "DBPort": dbPort, "DBUser": dbUser, "DBPassword": dbPassword, "DBName": DBName}).Info("Starting embedded PostgreSQL server")
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

		log.WithFields(log.Fields{"DBHost": dbHost, "DBPort": dbPort, "DBUser": dbUser, "DBPassword": dbPassword, "DBName": DBName}).Info("Connecting to PostgreSQL server")
		coloniesDB := postgresql.CreatePQDatabase(dbHost, dbPort, dbUser, dbPassword, DBName, DBPrefix)
		err = coloniesDB.Connect()
		CheckError(err)

		log.Info("Initialize a Colonies PostgreSQL database")
		err = coloniesDB.Initialize()
		CheckError(err)

		keychain, err := security.CreateKeychain(".colonies")
		CheckError(err)

		serverID := os.Getenv("COLONIES_SERVERID")
		serverPrvKey := os.Getenv("COLONIES_SERVERPRVKEY")
		log.WithFields(log.Fields{"ServerId": serverID, "ServerPrvKey": serverPrvKey}).Info("Adding a ServerId to keychain")
		err = keychain.AddPrvKey(serverID, serverPrvKey)
		CheckError(err)

		colonyID := os.Getenv("COLONIES_COLONYID")
		colonyPrvKey := os.Getenv("COLONIES_COLONYPRVKEY")
		log.WithFields(log.Fields{"ColonyId": colonyID, "ColonyPrvKey": colonyPrvKey}).Info("Adding a ColonyId to keychain")
		err = keychain.AddPrvKey(colonyID, colonyPrvKey)
		CheckError(err)

		runtimeID := os.Getenv("COLONIES_RUNTIMEID")
		runtimePrvKey := os.Getenv("COLONIES_RUNTIMEPRVKEY")
		log.WithFields(log.Fields{"RuntimeId": runtimeID, "RuntimePrvKey": runtimePrvKey}).Info("Adding a RuntimeId to keychain")
		err = keychain.AddPrvKey(runtimeID, runtimePrvKey)
		CheckError(err)

		coloniesServerPort, err := strconv.Atoi(os.Getenv("COLONIES_SERVERPORT"))
		CheckError(err)

		coloniesServer := server.CreateColoniesServer(coloniesDB, coloniesServerPort, serverID, false, "", "", true)
		go coloniesServer.ServeForever()

		coloniesServerHost := os.Getenv("COLONIES_SERVERHOST")
		log.WithFields(log.Fields{"ColoniesServerHost": coloniesServerHost, "ColoniesServerPort": coloniesServerPort}).Info("Connecting to Colonies server")
		client := client.CreateColoniesClient(coloniesServerHost, coloniesServerPort, true, false)

		log.WithFields(log.Fields{"ColonyID": colonyID}).Info("Registering a new Colony")
		colony := core.CreateColony(colonyID, "dev")
		_, err = client.AddColony(colony, serverPrvKey)
		CheckError(err)

		runtimeType := os.Getenv("COLONIES_RUNTIMETYPE")
		runtimeGroup := os.Getenv("COLONIES_RUNTIMEGROUP")
		log.WithFields(log.Fields{"RuntimeID": runtimeID, "RuntimeType": runtimeType, "RuntimeGroup": runtimeGroup}).Info("Registering a new Runtime")

		runtime := core.CreateRuntime(runtimeID, runtimeType, "dev_runtime", runtimeGroup, colonyID, "", 1, 0, "", 0, time.Now(), time.Now())
		_, err = client.AddRuntime(runtime, colonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"RuntimeID": runtimeID}).Info("Approving Runtime")
		log.Info("Approving CLI runtime")
		err = client.ApproveRuntime(runtimeID, colonyPrvKey)
		CheckError(err)

		wait := make(chan bool)

		log.Info("Successfully started Colonies development server")

		log.Info("Add the following environmental variables to your shell:")
		envStr := "export LANG=en_US.UTF-8\n"
		envStr += "export LANGUAGE=en_US.UTF-8\n"
		envStr += "export LC_ALL=en_US.UTF-8\n"
		envStr += "export LC_CTYPE=UTF-8\n"
		envStr += "export TZ=Europe/Stockholm\n"
		envStr += "export COLONIES_SERVERPROTOCOL=\"http\"\n"
		envStr += "export COLONIES_SERVERHOST=\"" + coloniesServerHost + " \"\n"
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
