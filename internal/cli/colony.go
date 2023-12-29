package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	colonyCmd.AddCommand(addColonyCmd)
	colonyCmd.AddCommand(removeColonyCmd)
	colonyCmd.AddCommand(lsColoniesCmd)
	colonyCmd.AddCommand(colonyStatsCmd)
	rootCmd.AddCommand(colonyCmd)

	colonyCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", DefaultServerHost, "Server host")
	colonyCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	addColonyCmd.Flags().StringVarP(&TargetColonyID, "colonyid", "", "", "Colony Id")
	addColonyCmd.MarkFlagRequired("colonyid")
	addColonyCmd.Flags().StringVarP(&TargetColonyName, "name", "", "", "Unique name of the Colony")
	addColonyCmd.MarkFlagRequired("name")

	removeColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	removeColonyCmd.Flags().StringVarP(&TargetColonyName, "name", "", "", "Colony name")
	removeColonyCmd.MarkFlagRequired("colonyid")

	lsColoniesCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	lsColoniesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
}

var colonyCmd = &cobra.Command{
	Use:   "colony",
	Short: "Manage colonies",
	Long:  "Manage colonies",
}

var addColonyCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new Colony",
	Long:  "Add a new Colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if TargetColonyName == "" {
			CheckError(errors.New("Target Colony name must be specifed"))
		}

		if TargetColonyID == "" {
			CheckError(errors.New("Target Colony Id must be specifed"))
		}

		colony := &core.Colony{Name: TargetColonyName, ID: TargetColonyID}

		addedColony, err := client.AddColony(colony, ServerPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ColonyName": TargetColonyName, "ColonyID": addedColony.ID}).Info("Colony added")
	},
}

var removeColonyCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a Colony",
	Long:  "Remove a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if TargetColonyName == "" {
			CheckError(errors.New("Colony name  not specified"))
		}

		err := client.RemoveColony(TargetColonyName, ServerPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ColonyName": TargetColonyName}).Info("Colony removed")
	},
}

var lsColoniesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all Colonies",
	Long:  "List all Colonies",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		coloniesFromServer, err := client.GetColonies(ServerPrvKey)
		CheckError(err)

		if len(coloniesFromServer) == 0 {
			log.Info("No Colonies found")
			os.Exit(0)
		}

		if JSON {
			jsonString, err := core.ConvertColonyArrayToJSON(coloniesFromServer)
			CheckError(err)
			fmt.Println(jsonString)
			os.Exit(0)
		}

		printColonyTable(coloniesFromServer)
	},
}

var colonyStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics about a Colony",
	Long:  "Show statistics about a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		stat, err := client.ColonyStatistics(ColonyName, PrvKey)
		CheckError(err)

		printColonyStatTable(stat)
	},
}
