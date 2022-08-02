package cli

import (
	"errors"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/monitoring"
	"github.com/colonyos/colonies/pkg/security"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	monitoringCmd.AddCommand(monitoringStartCmd)
	rootCmd.AddCommand(monitoringCmd)
}

var monitoringCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Manage Prometheus monitoring",
	Long:  "Manage Prometheus monitoring",
}

var monitoringStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Register and start a worker",
	Long:  "Register and start a worker",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		if ServerID == "" {
			ServerID = os.Getenv("COLONIES_SERVERID")
		}
		if ServerID == "" {
			CheckError(errors.New("Unknown Server Id"))
		}

		ServerPrvKey = os.Getenv("COLONIES_SERVERPRVKEY")
		if ServerPrvKey == "" {
			keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
			CheckError(err)
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		MonitorPortEnvStr := os.Getenv("COLONIES_MONITORPORT")
		if MonitorPortEnvStr == "" {
			CheckError(errors.New("COLONIES_MONITORPORT environmental variable not set"))
		}
		MonitorPort, err := strconv.Atoi(MonitorPortEnvStr)
		CheckError(err)

		IntervallEnvStr := os.Getenv("COLONIES_MONITORINTERVALL")
		if IntervallEnvStr == "" {
			CheckError(errors.New("COLONIES_MONITORINTERVALL environmental variable not set"))
		}
		MonitorIntervall, err = strconv.Atoi(IntervallEnvStr)
		CheckError(err)

		log.WithFields(log.Fields{
			"Port":               MonitorPort,
			"ColoniesServerHost": ServerHost,
			"ColoniesServerPort": ServerPort,
			"PullIntervall":      MonitorIntervall,
			"Insecure":           Insecure}).
			Info("Starting Prometheus monitoring server")
		monitoring.CreateMonitoringServer(MonitorPort, ServerHost, ServerPort, Insecure, SkipTLSVerify, ServerPrvKey, MonitorIntervall)

		wait := make(chan struct{})
		<-wait
	},
}
