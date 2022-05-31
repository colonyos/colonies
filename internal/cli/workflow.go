package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
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
	rootCmd.AddCommand(workflowCmd)

	submitWorkflowCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	submitWorkflowCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	submitWorkflowCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony workflow")
	submitWorkflowCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	submitWorkflowCmd.Flags().BoolVarP(&Wait, "wait", "", false, "Colony Id")
	submitWorkflowCmd.MarkFlagRequired("spec")

	listWaitingWorkflowsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listWaitingWorkflowsCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listWaitingWorkflowsCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listWaitingWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	listRunningWorkflowsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listRunningWorkflowsCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listRunningWorkflowsCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listRunningWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	listSuccessfulWorkflowsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listSuccessfulWorkflowsCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listSuccessfulWorkflowsCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listSuccessfulWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	listFailedWorkflowsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listFailedWorkflowsCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listFailedWorkflowsCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listFailedWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	getWorkflowCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	getWorkflowCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
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
	Short: "Submit a workflow to a Colony",
	Long:  "Submit a workflow to a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		workflowSpec, err := core.ConvertJSONToWorkflowSpec(string(jsonSpecBytes))
		CheckError(err)

		if workflowSpec.ColonyID == "" {
			if ColonyID == "" {
				ColonyID = os.Getenv("COLONIES_COLONYID")
			}
			if ColonyID == "" {
				CheckError(errors.New("Unknown Colony Id, please set COLONYID env variable or specify ColonyID in JSON file"))
			}

			workflowSpec.ColonyID = ColonyID
		}

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		graph, err := client.SubmitWorkflowSpec(workflowSpec, RuntimePrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"WorkflowID": graph.ID}).Info("Workflow submitted")
	},
}

var listWaitingWorkflowsCmd = &cobra.Command{
	Use:   "psw",
	Short: "List all waiting workflows",
	Long:  "List all waiting workflows",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		graphs, err := client.GetWaitingProcessGraphs(ColonyID, Count, RuntimePrvKey)
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

var listRunningWorkflowsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all running workflows",
	Long:  "List all running workflows",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		graphs, err := client.GetRunningProcessGraphs(ColonyID, Count, RuntimePrvKey)
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		graphs, err := client.GetSuccessfulProcessGraphs(ColonyID, Count, RuntimePrvKey)
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		graphs, err := client.GetFailedProcessGraphs(ColonyID, Count, RuntimePrvKey)
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

var getWorkflowCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a workflow",
	Long:  "Get info about a workflow",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		graph, err := client.GetProcessGraph(WorkflowID, RuntimePrvKey)
		if err != nil {
			log.WithFields(log.Fields{"WorkflowID": WorkflowID, "Error": err}).Error("Workflow not found")
			os.Exit(-1)
		}

		if JSON {
			fmt.Println(graph.ToJSON())
			os.Exit(0)
		}

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
			process, err := client.GetProcess(processID, RuntimePrvKey)
			CheckError(err)

			procCmd := process.ProcessSpec.Cmd
			if procCmd == "" {
				procCmd = "None"
			}

			procArgs := ""
			for _, procArg := range process.ProcessSpec.Args {
				procArgs += procArg + " "
			}
			if procArgs == "" {
				procArgs = "None"
			}

			dependencies := ""
			for _, dependency := range process.ProcessSpec.Conditions.Dependencies {
				dependencies += dependency + " "
			}
			if dependencies == "" {
				dependencies = "None"
			}

			processData := [][]string{
				[]string{"Name", process.ProcessSpec.Name},
				[]string{"ProcessID", process.ID},
				[]string{"RuntimeType", process.ProcessSpec.Conditions.RuntimeType},
				[]string{"Cmd", procCmd},
				[]string{"Args", procArgs},
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
	},
}
