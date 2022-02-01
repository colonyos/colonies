package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const DBName = "postgres"
const DBPrefix = "PROD_"
const KEYCHAIN_PATH = ".colonies"
const TimeLayout = "2006-01-02 15:04:05"

var Verbose bool
var DBHost string
var DBPort int
var DBUser string
var DBPassword string
var BindAddr string
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
var RuntimeID string
var RuntimePrvKey string
var ColonyPrvKey string
var ColonyID string
var ProcessID string
var Key string
var Value string
var AttributeID string
var JSON bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
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
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
