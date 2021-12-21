package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const DBName = "postgres"
const DBPrefix = "PROD_"
const KEYCHAIN_PATH = ".colonies"

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
var RootPassword string
var SpecFile string
var Count int
var ID string
var PrvKey string
var ComputerID string
var ComputerPrvKey string
var ColonyPrvKey string
var ColonyID string

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
		fmt.Println(err)
		os.Exit(-1)
	}
}
