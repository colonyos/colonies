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
	colonyCmd.AddCommand(chColonyIDCmd)
	colonyCmd.AddCommand(removeColonyCmd)
	colonyCmd.AddCommand(lsColoniesCmd)
	colonyCmd.AddCommand(colonyStatsCmd)
	colonyCmd.AddCommand(checkColonyCmd)
	rootCmd.AddCommand(colonyCmd)

	colonyCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", DefaultServerHost, "Server host")
	colonyCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	addColonyCmd.Flags().StringVarP(&TargetColonyID, "colonyid", "", "", "Colony Id")
	addColonyCmd.MarkFlagRequired("colonyid")
	addColonyCmd.Flags().StringVarP(&TargetColonyName, "name", "", "", "Unique name of the Colony")
	addColonyCmd.MarkFlagRequired("name")

	chColonyIDCmd.Flags().StringVarP(&TargetColonyID, "colonyid", "", "", "Colony Id")
	chColonyIDCmd.MarkFlagRequired("colonyid")

	removeColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	removeColonyCmd.Flags().StringVarP(&TargetColonyName, "name", "", "", "Colony name")
	removeColonyCmd.MarkFlagRequired("colonyid")

	lsColoniesCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	lsColoniesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	checkColonyCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colonies server private key")
	checkColonyCmd.Flags().StringVarP(&TargetColonyName, "name", "", "", "Unique name of the Colony")
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

var chColonyIDCmd = &cobra.Command{
	Use:   "chid",
	Short: "Change colony Id",
	Long:  "Change colony Id",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if len(TargetColonyID) != 64 {
			CheckError(errors.New("Invalid colony Id length"))
		}

		err := client.ChangeColonyID(ColonyName, TargetColonyID, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ColonyName": ColonyName,
			"ColonyId":   TargetColonyID}).
			Info("Changed colony Id")
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

		if JSON {
			jsonString, err := core.ConvertColonyArrayToJSON(coloniesFromServer)
			if jsonString == "null" {
				jsonString = "[]"
			}
			CheckError(err)
			fmt.Println(jsonString)
			os.Exit(0)
		}

		if len(coloniesFromServer) == 0 {
			log.Info("No Colonies found")
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

var checkColonyCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if a Colony exists",
	Long:  "Check if a Colony exists",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if TargetColonyName == "" {
			CheckError(errors.New("Colony name  not specified"))
		}

		_, err := client.GetColonyByName(TargetColonyName, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ColonyName": TargetColonyName}).Info("Colony exists")
	},
}
