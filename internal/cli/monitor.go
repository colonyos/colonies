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
	Short: "Start a monitoring server",
	Long:  "Start a monitoring server",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		if ServerID == "" {
			ServerID = os.Getenv("COLONIES_SERVER_ID")
		}
		if ServerID == "" {
			CheckError(errors.New("Unknown Server Id"))
		}

		ServerPrvKey = os.Getenv("COLONIES_SERVER_PRVKEY")
		if ServerPrvKey == "" {
			keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
			CheckError(err)
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		MonitorPortEnvStr := os.Getenv("COLONIES_MONITOR_PORT")
		if MonitorPortEnvStr == "" {
			CheckError(errors.New("COLONIES_MONITORPORT environmental variable not set"))
		}
		MonitorPort, err := strconv.Atoi(MonitorPortEnvStr)
		CheckError(err)

		IntervalEnvStr := os.Getenv("COLONIES_MONITOR_INTERVAL")
		if IntervalEnvStr == "" {
			CheckError(errors.New("COLONIES_MONITOR_INTERVAL environmental variable not set"))
		}
		MonitorInterval, err = strconv.Atoi(IntervalEnvStr)
		CheckError(err)

		log.WithFields(log.Fields{
			"Port":               MonitorPort,
			"ColoniesServerHost": ServerHost,
			"ColoniesServerPort": ServerPort,
			"PullInterval":       MonitorInterval,
			"Insecure":           Insecure}).
			Info("Starting Prometheus monitoring server")
		monitoring.CreateMonitoringServer(MonitorPort, ServerHost, ServerPort, Insecure, SkipTLSVerify, ServerPrvKey, MonitorInterval)

		wait := make(chan struct{})
		<-wait
	},
}
