package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/kataras/tablewriter"
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

	colonyStatsCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	colonyStatsCmd.Flags().StringVarP(&TargetColonyName, "name", "", "", "Colony name")
	colonyStatsCmd.MarkFlagRequired("name")
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
			data = append(data, []string{colony.Name})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name"})

		for _, v := range data {
			table.Append(v)
		}

		table.Render()
	},
}

var colonyStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics about a Colony",
	Long:  "Show statistics about a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if TargetColonyName == "" {
			CheckError(errors.New("Colony name not specified"))
		}

		stat, err := client.ColonyStatistics(TargetColonyName, PrvKey)
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
