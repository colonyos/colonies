package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const DBPrefix = "PROD_"
const KEYCHAIN_PATH = ".colonies"
const TimeLayout = "2006-01-02 15:04:05"
const DefaultDBHost = "localhost"
const DefaultDBPort = 5432
const DefaultServerHost = "localhost"
const MaxAttributeLength = 30

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
var ExecutorName string
var ExecutorID string
var ExecutorType string
var ExecutorPrvKey string
var TargetExecutorID string
var TargetExecutorType string
var TargetExecutorName string
var WorkflowID string
var ColonyPrvKey string
var ColonyID string
var ProcessID string
var Key string
var Value string
var AttributeID string
var JSON bool
var Wait bool
var PrintOutput bool
var Full bool
var LogDir string
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
var Latest bool
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
var ExclusiveAssign bool
var Approve bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&Insecure, "insecure", "", false, "Disable TLS and use HTTP")
	rootCmd.PersistentFlags().BoolVarP(&SkipTLSVerify, "skip-tls-verify", "", false, "Skip TLS certificate verification")
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
