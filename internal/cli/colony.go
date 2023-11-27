package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	colonyCmd.AddCommand(addColonyCmd)
	colonyCmd.AddCommand(removeColonyCmd)
	colonyCmd.AddCommand(renameColonyCmd)
	colonyCmd.AddCommand(lsColoniesCmd)
	colonyCmd.AddCommand(colonyStatsCmd)
	rootCmd.AddCommand(colonyCmd)

	colonyCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", DefaultServerHost, "Server host")
	colonyCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	addColonyCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	addColonyCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony")
	addColonyCmd.MarkFlagRequired("spec")

	removeColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	removeColonyCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	removeColonyCmd.MarkFlagRequired("colonyid")

	renameColonyCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	renameColonyCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	renameColonyCmd.Flags().StringVarP(&ColonyName, "name", "", "", "New Colony name")
	renameColonyCmd.MarkFlagRequired("name")

	lsColoniesCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	lsColoniesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	colonyStatsCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	colonyStatsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
}

var colonyCmd = &cobra.Command{
	Use:   "colony",
	Short: "Manage colonies",
	Long:  "Manage colonies",
}

var addColonyCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new colony",
	Long:  "Add a new colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		colony, err := core.ConvertJSONToColony(string(jsonSpecBytes))
		CheckError(err)

		crypto := crypto.CreateCrypto()

		var prvKey string
		if ColonyPrvKey != "" {
			prvKey = ColonyPrvKey
			if len(prvKey) != 64 {
				CheckError(errors.New("Invalid private key length"))
			}
		} else {
			prvKey, err = crypto.GeneratePrivateKey()
			CheckError(errors.New("No Colony private key specified"))
		}

		colonyID, err := crypto.GenerateID(prvKey)
		CheckError(err)
		colony.SetID(colonyID)

		addedColony, err := client.AddColony(colony, ServerPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ColonyID": addedColony.ID}).Info("Colony added")
	},
}

var removeColonyCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a colony",
	Long:  "Remove a colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.DeleteColony(ColonyID, ServerPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ColonyID": ColonyID}).Info("Colony removed")
	},
}

var renameColonyCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename a colony",
	Long:  "Rename a colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if ColonyName == "" {
			CheckError(errors.New("Invalid Colony name"))
		}

		err := client.RenameColony(ColonyID, ColonyName, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ColonyID": ColonyID, "Name": ColonyName}).Info("Colony renamed")
	},
}

var lsColoniesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all colonies",
	Long:  "List all colonies",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		coloniesFromServer, err := client.GetColonies(ServerPrvKey)
		CheckError(err)

		if len(coloniesFromServer) == 0 {
			log.Warning("No colonies found")
			os.Exit(0)
		}

		if JSON {
			jsonString, err := core.ConvertColonyArrayToJSON(coloniesFromServer)
			CheckError(err)
			fmt.Println(jsonString)
			os.Exit(0)
		}

		var data [][]string
		for _, colony := range coloniesFromServer {
			data = append(data, []string{colony.ID, colony.Name})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name"})

		for _, v := range data {
			table.Append(v)
		}

		table.Render()
	},
}

var colonyStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics about a colony",
	Long:  "Show statistics about a colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		stat, err := client.ColonyStatistics(ColonyID, PrvKey)
		CheckError(err)

		fmt.Println("Process statistics:")
		specData := [][]string{
			[]string{"Executors", strconv.Itoa(stat.Executors)},
			[]string{"Waiting processes", strconv.Itoa(stat.WaitingProcesses)},
			[]string{"Running processes", strconv.Itoa(stat.RunningProcesses)},
			[]string{"Successful processes", strconv.Itoa(stat.SuccessfulProcesses)},
			[]string{"Failed processes", strconv.Itoa(stat.FailedProcesses)},
			[]string{"Waiting workflows", strconv.Itoa(stat.WaitingWorkflows)},
			[]string{"Running workflows ", strconv.Itoa(stat.RunningWorkflows)},
			[]string{"Successful workflows", strconv.Itoa(stat.SuccessfulWorkflows)},
			[]string{"Failed workflows", strconv.Itoa(stat.FailedWorkflows)},
		}
		specTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range specData {
			specTable.Append(v)
		}
		specTable.SetAlignment(tablewriter.ALIGN_LEFT)
		specTable.Render()
	},
}
