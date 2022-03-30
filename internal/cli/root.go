package cli

import (
	"fmt"
	"os"

	"github.com/colonyos/colonies/pkg/build"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const DBPrefix = "PROD_"
const KEYCHAIN_PATH = ".colonies"
const TimeLayout = "2006-01-02 15:04:05"
const DefaultDBHost = "localhost"
const DefaultDBPort = 5432
const DefaultServerHost = "localhost"
const DefaultServerPort = 50080
const MaxAttributeLength = 30

var DBName = "postgres"
var Verbose bool
var DBHost string
var DBPort int
var DBUser string
var DBPassword string
var BindAddr string
var TLS bool
var TLSCert string
var TLSKey string
var ServerHost string
var ServerPort int
var ServerID string
var ServerPrvKey string
var SpecFile string
var Count int
var ID string
var PrvKey string
var RuntimeName string
var RuntimeType string
var RuntimeID string
var RuntimePrvKey string
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

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&TLS, "tls", "", true, "Enable/disable TLS")
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
