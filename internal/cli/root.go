package cli

import (
	"fmt"
	"os"

	"github.com/colonyos/colonies/pkg/build"
	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
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
var MonitorIntervall int
var ServerID string
var ServerPrvKey string
var SpecFile string
var Count int
var ID string
var PrvKey string
var RuntimeName string
var RuntimeID string
var RuntimeType string
var RuntimePrvKey string
var TargetRuntimeID string
var TargetRuntimeType string
var TargetRuntimeName string
var WorkflowID string
var CPU string
var Cores int
var Mem int
var GPU string
var GPUs int
var ColonyPrvKey string
var ColonyID string
var ProcessID string
var Key string
var Value string
var AttributeID string
var JSON bool
var Wait bool
var Output bool
var Full bool
var LogDir string
var GeneratorID string
var GeneratorName string
var GeneratorTrigger int
var GeneratorTimeout int
var Func string
var Data string
var Args []string
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
var CronIntervall int
var CronRandom bool

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

func CheckError(err error) {
	if err != nil {
		log.WithFields(log.Fields{"err": err, "BuildVersion": build.BuildVersion, "BuildTime": build.BuildTime}).Error(err.Error())
		os.Exit(-1)
	}
}

func Args2String(args []string) string {
	if len(args) == 0 {
		return ""
	}

	str := ""
	for _, arg := range args {
		str += arg + " "
	}

	return str[0 : len(str)-1]
}

func State2String(state int) string {
	var stateStr string
	switch state {
	case core.WAITING:
		stateStr = "Waiting"
	case core.RUNNING:
		stateStr = "Running"
	case core.SUCCESS:
		stateStr = "Successful"
	case core.FAILED:
		stateStr = "Failed"
	default:
		stateStr = "Unkown"
	}

	return stateStr
}
