package cli

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
)

const DBPrefix = "PROD_"
const KEYCHAIN_PATH = ".colonies"
const TimeLayout = "2006-01-02 15:04:05"
const DefaultDBHost = "localhost"
const DefaultDBPort = 5432
const DefaultServerHost = "localhost"
const MaxAttributeLength = 30
const MaxArgLength = 20
const MaxArgInfoLength = 50

var mutex sync.Mutex

var DBName = "postgres"
var Verbose bool
var DBHost string
var DBPort int
var DBUser string
var DBPassword string
var BindAddr string
var Insecure bool
var SkipTLSVerify bool
var UseTLS bool
var TLSCert string
var TLSKey string
var ServerHost string
var ServerPort int
var MonitorPort int
var MonitorInterval int
var ServerID string
var ServerPrvKey string
var SpecFile string
var Count int
var ID string
var PrvKey string
var ExecutorID string
var ExecutorName string
var ExecutorType string
var FunctionID string
var TargetColonyID string
var TargetColonyName string
var NewColonyName string
var TargetExecutorID string
var TargetExecutorType string
var TargetExecutorName string
var WorkflowID string
var ColonyPrvKey string
var ColonyName string
var ProcessID string
var Key string
var Value string
var AttributeID string
var JSON bool
var Wait bool
var PrintOutput bool
var Full bool
var GeneratorID string
var GeneratorName string
var GeneratorTrigger int
var GeneratorTimeout int
var GeneratorCheckerPeriod int
var FuncName string
var Arg string
var Args []string
var Output []string
var Errors []string
var Env []string
var MaxWaitTime int
var MaxExecTime int
var MaxRetries int
var EtcdName string
var EtcdHost string
var EtcdClientPort int
var EtcdPeerPort int
var EtcdCluster []string
var EtcdDataDir string
var RelayPort int
var Timeout int
var CronID string
var CronName string
var CronExpr string
var CronInterval int
var CronCheckerPeriod int
var CronRandom bool
var WaitForPrevProcessGraph bool
var Long float64
var Lat float64
var AllowExecutorReregister bool
var ExclusiveAssign bool
var Approve bool
var Waiting bool
var Successful bool
var Failed bool
var TimescaleDB bool
var LogMsg string
var Since int64
var Follow bool
var SyncDir string
var StorageDriver string
var Label string
var Dry bool
var KeepLocal bool
var Yes bool
var Filename string
var FileID string
var DownloadDir string
var SnapshotID string
var SnapshotName string
var KwArgs []string
var Snapshots []string
var Retention bool
var RetentionPolicy int64
var UserPrvKey string
var UserID string
var Username string
var Email string
var Phone string
var ShowIDs bool
var InitDB bool
var SyncPlans bool
var Quite bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Verbose (debugging)")
	rootCmd.PersistentFlags().BoolVarP(&Insecure, "insecure", "", false, "Disable TLS and use HTTP")
	rootCmd.PersistentFlags().BoolVarP(&SkipTLSVerify, "skip-tls-verify", "", false, "Skip TLS certificate verification")

	rootCmd.AddCommand(configCmd)
}

var rootCmd = &cobra.Command{
	Use:   "colonies",
	Short: "Colonies CLI tool",
	Long:  "Colonies CLI tool",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show currently used configuration",
	Long:  "Show currently used configuration",
	Run: func(cmd *cobra.Command, args []string) {
		setup()
		printConfigTable()
	},
}
