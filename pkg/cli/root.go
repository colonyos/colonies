package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Verbose bool
var DBHost string
var DBPort int
var DBUser string
var DBPassword string
var DBName string
var BindAddr string
var ServerAddr string

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
