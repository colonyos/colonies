package cli

import (
	"colonies/internal/logging"
	"colonies/pkg/database/postgresql"
	"colonies/pkg/server"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
	serverCmd.AddCommand(serverStartCmd)
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().StringVarP(&DBHost, "dbhost", "", "localhost", "Colonies database host")
	serverCmd.MarkPersistentFlagRequired("dbhost")
	serverCmd.PersistentFlags().IntVarP(&DBPort, "dbport", "", 5432, "Colonies database port")
	serverCmd.MarkPersistentFlagRequired("dbport")
	serverCmd.PersistentFlags().StringVarP(&DBUser, "dbuser", "", "", "Colonies database user")
	serverCmd.MarkPersistentFlagRequired("dbuser")
	serverCmd.PersistentFlags().StringVarP(&DBPassword, "dbpassword", "", "", "Colonies database password")
	serverCmd.MarkPersistentFlagRequired("dbpassword")
	serverCmd.PersistentFlags().StringVarP(&TLSCert, "tlscert", "", "", "TLS certificate")
	serverCmd.MarkPersistentFlagRequired("tlscert")
	serverCmd.PersistentFlags().StringVarP(&TLSKey, "tlskey", "", "", "TLS key")
	serverCmd.MarkPersistentFlagRequired("tlskey")
	serverCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Server HTTP port")
	serverCmd.MarkPersistentFlagRequired("port")
	serverCmd.PersistentFlags().StringVarP(&RootPassword, "rootpassword", "", "", "Root password to the Colonies server")
	serverCmd.MarkPersistentFlagRequired("rootpassword")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage a Colonies server",
	Long:  "Manage a Colonies server",
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a Colonies server",
	Long:  "Start a Colonies server",
	Run: func(cmd *cobra.Command, args []string) {
		db := postgresql.CreatePQDatabase(DBHost, DBPort, DBUser, DBPassword, DBName, DBPrefix)
		err := db.Connect()
		if err != nil {
			fmt.Println("Failed to connect to database")
			os.Exit(-1)
		}
		logging.Log().Info("Connecting to Colonies database, host: " + DBHost + ", port: " + strconv.Itoa(DBPort) + ", user: " + DBUser + ", password: " + "******************, name: " + DBName + ". prefix: " + DBPrefix)
		server := server.CreateColoniesServer(db, ServerPort, RootPassword, TLSKey, TLSCert, Verbose)
		server.ServeForever()
	},
}
