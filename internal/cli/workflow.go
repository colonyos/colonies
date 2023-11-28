package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	workflowCmd.AddCommand(submitWorkflowCmd)
	workflowCmd.AddCommand(listWaitingWorkflowsCmd)
	workflowCmd.AddCommand(listRunningWorkflowsCmd)
	workflowCmd.AddCommand(listSuccessfulWorkflowsCmd)
	workflowCmd.AddCommand(listFailedWorkflowsCmd)
	workflowCmd.AddCommand(getWorkflowCmd)
	workflowCmd.AddCommand(deleteWorkflowCmd)
	workflowCmd.AddCommand(deleteAllWorkflowsCmd)
	rootCmd.AddCommand(workflowCmd)

	submitWorkflowCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	submitWorkflowCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony workflow")
	submitWorkflowCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	submitWorkflowCmd.Flags().BoolVarP(&Wait, "wait", "", false, "Colony Id")
	submitWorkflowCmd.MarkFlagRequired("spec")

	listWaitingWorkflowsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listWaitingWorkflowsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listWaitingWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	listRunningWorkflowsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listRunningWorkflowsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listRunningWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	listSuccessfulWorkflowsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listSuccessfulWorkflowsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listSuccessfulWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	listFailedWorkflowsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listFailedWorkflowsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listFailedWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	deleteWorkflowCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	deleteWorkflowCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	deleteWorkflowCmd.Flags().StringVarP(&WorkflowID, "workflowid", "", "", "Workflow Id")
	deleteWorkflowCmd.MarkFlagRequired("processid")

	deleteAllWorkflowsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	deleteAllWorkflowsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	deleteAllWorkflowsCmd.Flags().BoolVarP(&Waiting, "waiting", "", false, "Delete all waiting processes")
	deleteAllWorkflowsCmd.Flags().BoolVarP(&Successful, "successful", "", false, "Delete all successful processes")
	deleteAllWorkflowsCmd.Flags().BoolVarP(&Failed, "failed", "", false, "Delete all failed processes")

	getWorkflowCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	getWorkflowCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	getWorkflowCmd.Flags().StringVarP(&WorkflowID, "workflowid", "", "", "Workflow Id")
	getWorkflowCmd.MarkFlagRequired("workflowid")
	getWorkflowCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
}

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage workflows",
	Long:  "Manage workflows",
}

var submitWorkflowCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a workflow",
	Long:  "Submit a workflow",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		jsonStr := "{\"functionspecs\":" + string(jsonSpecBytes) + "}"
		workflowSpec, err := core.ConvertJSONToWorkflowSpec(jsonStr)
		CheckJSONParseErr(err, string(jsonSpecBytes))

		if workflowSpec.ColonyName == "" {
			if ColonyName == "" {
				ColonyName = os.Getenv("COLONIES_COLONY_NAME")
			}
			if ColonyName == "" {
				CheckError(errors.New("Unknown Colony Id, please export COLONIES_COLONY_NAME variable or specify ColonyName in JSON file"))
			}

			workflowSpec.ColonyName = ColonyName
		}

		graph, err := client.SubmitWorkflowSpec(workflowSpec, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"WorkflowID": graph.ID}).Info("Workflow submitted")
	},
}

var listWaitingWorkflowsCmd = &cobra.Command{
	Use:   "psw",
	Short: "List all waiting workflows",
	Long:  "List all waiting workflows",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		graphs, err := client.GetWaitingProcessGraphs(ColonyID, Count, PrvKey)
		CheckError(err)

		if len(graphs) == 0 {
			log.Warning("No waiting workflows found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, graph := range graphs {
				data = append(data, []string{graph.ID, graph.SubmissionTime.Format(TimeLayout)})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Submission Time"})
			for _, v := range data {
				table.Append(v)
			}
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
		}

	},
}

var deleteWorkflowCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a workflow",
	Long:  "Delete a workflow",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.DeleteProcessGraph(WorkflowID, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"WorkflowID": WorkflowID}).Info("ProcessGraph deleted")
	},
}

var deleteAllWorkflowsCmd = &cobra.Command{
	Use:   "deleteall",
	Short: "Delete all workflows",
	Long:  "Delete all workflows",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		counter := 0
		state := ""
		if Waiting {
			counter++
			state = "waiting"
		}
		if Successful {
			counter++
			state = "successful"
		}
		if Failed {
			counter++
			state = "failed"
		}

		if counter > 1 {
			CheckError(errors.New("Invalid flags, select --waiting, --successful or --failed"))
		}

		if counter == 0 {
			state = "all"
		}

		fmt.Print("WARNING!!! Are you sure you want to delete " + state + " workflows in the Colony This operation cannot be undone! (YES,no): ")

		var err error
		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')
		if reply == "YES\n" {
			if state == "all" {
				err = client.DeleteAllProcessGraphs(ColonyID, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyID": ColonyID}).Info("Deleting all workflows in Colony")
			} else if Waiting {
				err = client.DeleteAllProcessGraphsWithState(ColonyID, core.WAITING, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyID": ColonyID}).Info("Deleting all waiting workflows in Colony")
			} else if Successful {
				err = client.DeleteAllProcessGraphsWithState(ColonyID, core.SUCCESS, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyID": ColonyID}).Info("Deleting all successful workflows in Colony")
			} else if Failed {
				err = client.DeleteAllProcessGraphsWithState(ColonyID, core.FAILED, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyID": ColonyID}).Info("Deleting all failed workflows in Colony")
			}

		} else {
			log.Info("Aborting ...")
		}
	},
}

var listRunningWorkflowsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all running workflows",
	Long:  "List all running workflows",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		graphs, err := client.GetRunningProcessGraphs(ColonyID, Count, PrvKey)
		CheckError(err)

		if len(graphs) == 0 {
			log.Warning("No running workflows found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, graph := range graphs {
				data = append(data, []string{graph.ID, graph.SubmissionTime.Format(TimeLayout)})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Submission Time"})
			for _, v := range data {
				table.Append(v)
			}
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
		}

	},
}

var listSuccessfulWorkflowsCmd = &cobra.Command{
	Use:   "pss",
	Short: "List all successful workflows",
	Long:  "List all successful workflows",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		graphs, err := client.GetSuccessfulProcessGraphs(ColonyID, Count, PrvKey)
		CheckError(err)

		if len(graphs) == 0 {
			log.Warning("No successful workflows found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, graph := range graphs {
				data = append(data, []string{graph.ID, graph.EndTime.Format(TimeLayout)})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "End Time"})
			for _, v := range data {
				table.Append(v)
			}
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
		}

	},
}

var listFailedWorkflowsCmd = &cobra.Command{
	Use:   "psf",
	Short: "List all failed workflows",
	Long:  "List all failed workflows",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		graphs, err := client.GetFailedProcessGraphs(ColonyID, Count, PrvKey)
		CheckError(err)

		if len(graphs) == 0 {
			log.Warning("No failed workflows found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessGraphArrayToJSON(graphs)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, graph := range graphs {
				data = append(data, []string{graph.ID, graph.EndTime.Format(TimeLayout)})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "End Time"})
			for _, v := range data {
				table.Append(v)
			}
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
		}

	},
}

func printGraf(client *client.ColoniesClient, graph *core.ProcessGraph) {
	fmt.Println("Workflow:")
	workflowData := [][]string{
		[]string{"WorkflowID", graph.ID},
		[]string{"ColonyID", graph.ID},
		[]string{"State", State2String(graph.State)},
		[]string{"SubmissionTime", graph.SubmissionTime.Format(TimeLayout)},
		[]string{"StartTime", graph.StartTime.Format(TimeLayout)},
		[]string{"EndTime", graph.EndTime.Format(TimeLayout)},
	}
	workflowTable := tablewriter.NewWriter(os.Stdout)
	for _, v := range workflowData {
		workflowTable.Append(v)
	}
	workflowTable.SetAlignment(tablewriter.ALIGN_LEFT)
	workflowTable.Render()

	fmt.Println("\nProcesses:")
	for i, processID := range graph.ProcessIDs {
		process, err := client.GetProcess(processID, PrvKey)
		CheckError(err)

		f := process.FunctionSpec.FuncName
		if f == "" {
			f = "None"
		}

		procArgs := ""
		for _, procArg := range IfArr2StringArr(process.FunctionSpec.Args) {
			procArgs += procArg + " "
		}
		if procArgs == "" {
			procArgs = "None"
		}

		procKwArgs := ""
		for k, procKwArg := range IfMap2StringMap(process.FunctionSpec.KwArgs) {
			procKwArgs += k + ":" + procKwArg + " "
		}
		if procKwArgs == "" {
			procKwArgs = "None"
		}

		dependencies := ""
		for _, dependency := range process.FunctionSpec.Conditions.Dependencies {
			dependencies += dependency + " "
		}
		if dependencies == "" {
			dependencies = "None"
		}

		processData := [][]string{
			[]string{"NodeName", process.FunctionSpec.NodeName},
			[]string{"ProcessID", process.ID},
			[]string{"ExecutorType", process.FunctionSpec.Conditions.ExecutorType},
			[]string{"FuncName", f},
			[]string{"Args", procArgs},
			[]string{"KwArgs", procKwArgs},
			[]string{"State", State2String(process.State)},
			[]string{"WaitingForParents", strconv.FormatBool(process.WaitForParents)},
			[]string{"Dependencies", dependencies},
		}
		processTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range processData {
			processTable.Append(v)
		}
		processTable.SetAlignment(tablewriter.ALIGN_LEFT)
		processTable.Render()

		if i < len(graph.ProcessIDs)-1 {
			fmt.Println()
		}
	}
}

var getWorkflowCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a workflow",
	Long:  "Get info about a workflow",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		graph, err := client.GetProcessGraph(WorkflowID, PrvKey)
		if err != nil {
			log.WithFields(log.Fields{"WorkflowID": WorkflowID, "Error": err}).Error("Workflow not found")
			os.Exit(-1)
		}

		if JSON {
			fmt.Println(graph.ToJSON())
			os.Exit(0)
		}

		printGraf(client, graph)
	},
}
