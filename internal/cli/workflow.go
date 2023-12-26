package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
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
	workflowCmd.AddCommand(removeWorkflowCmd)
	workflowCmd.AddCommand(removeAllWorkflowsCmd)
	rootCmd.AddCommand(workflowCmd)

	submitWorkflowCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	submitWorkflowCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony workflow")
	submitWorkflowCmd.Flags().StringVarP(&ColonyName, "colonyid", "", "", "Colony Id")
	submitWorkflowCmd.Flags().BoolVarP(&Wait, "wait", "", false, "Colony Id")
	submitWorkflowCmd.MarkFlagRequired("spec")

	listWaitingWorkflowsCmd.Flags().StringVarP(&ColonyName, "colonyid", "", "", "Colony Id")
	listWaitingWorkflowsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listWaitingWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	listRunningWorkflowsCmd.Flags().StringVarP(&ColonyName, "colonyid", "", "", "Colony Id")
	listRunningWorkflowsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listRunningWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	listSuccessfulWorkflowsCmd.Flags().StringVarP(&ColonyName, "colonyid", "", "", "Colony Id")
	listSuccessfulWorkflowsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listSuccessfulWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	listFailedWorkflowsCmd.Flags().StringVarP(&ColonyName, "colonyid", "", "", "Colony Id")
	listFailedWorkflowsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listFailedWorkflowsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of workflows to list")

	removeWorkflowCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	removeWorkflowCmd.Flags().StringVarP(&ColonyName, "colonyid", "", "", "Colony Id")
	removeWorkflowCmd.Flags().StringVarP(&WorkflowID, "workflowid", "", "", "Workflow Id")
	removeWorkflowCmd.MarkFlagRequired("processid")

	removeAllWorkflowsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	removeAllWorkflowsCmd.Flags().StringVarP(&ColonyName, "colonyid", "", "", "Colony Id")
	removeAllWorkflowsCmd.Flags().BoolVarP(&Waiting, "waiting", "", false, "Remove all waiting processes")
	removeAllWorkflowsCmd.Flags().BoolVarP(&Successful, "successful", "", false, "Remove all successful processes")
	removeAllWorkflowsCmd.Flags().BoolVarP(&Failed, "failed", "", false, "Remove all failed processes")

	getWorkflowCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	getWorkflowCmd.Flags().StringVarP(&ColonyName, "colonyid", "", "", "Colony Id")
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
		if err != nil {
			if strings.Contains(err.Error(), "cannot unmarshal object into Go struct field WorkflowSpec.functionspecs of type []core.FunctionSpec") {
				_, err := core.ConvertJSONToFunctionSpec(string(jsonSpecBytes))
				if err == nil {
					CheckError(errors.New("It looks like you are trying to submit a function spec, try to use colonies function submit --spec instead"))
				}
			}
		}
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

		log.WithFields(log.Fields{"ColonyName": ColonyName, "WorkflowID": graph.ID}).Info("Workflow submitted")
	},
}

var removeWorkflowCmd = &cobra.Command{
	Use:   "remove",
	Short: "remove a workflow",
	Long:  "remove a workflow",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.RemoveProcessGraph(WorkflowID, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"WorkflowID": WorkflowID}).Info("Workflow removed")
	},
}

var removeAllWorkflowsCmd = &cobra.Command{
	Use:   "removeall",
	Short: "Remove all workflows",
	Long:  "Remove all workflows",
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

		fmt.Print("WARNING!!! Are you sure you want to remove " + state + " workflows in the Colony <" + ColonyName + ">. This operation cannot be undone! (YES,no): ")

		var err error
		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')
		if reply == "YES\n" {
			if state == "all" {
				err = client.RemoveAllProcessGraphs(ColonyName, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all workflows in Colony <" + ColonyName + ">")
			} else if Waiting {
				err = client.RemoveAllProcessGraphsWithState(ColonyName, core.WAITING, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all waiting workflows in Colony <" + ColonyName + ">")
			} else if Successful {
				err = client.RemoveAllProcessGraphsWithState(ColonyName, core.SUCCESS, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all successful workflows in Colony <" + ColonyName + ">")
			} else if Failed {
				err = client.RemoveAllProcessGraphsWithState(ColonyName, core.FAILED, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all failed workflows in Colony <" + ColonyName + ">")
			}
		} else {
			log.Info("Aborting ...")
		}
	},
}

var listWaitingWorkflowsCmd = &cobra.Command{
	Use:   "psw",
	Short: "List all waiting workflows",
	Long:  "List all waiting workflows",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		graphs, err := client.GetWaitingProcessGraphs(ColonyName, Count, PrvKey)
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

			printWorkflowTable(graphs, core.WAITING)
		}
	},
}

var listRunningWorkflowsCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all running workflows",
	Long:  "List all running workflows",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		graphs, err := client.GetRunningProcessGraphs(ColonyName, Count, PrvKey)
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

			printWorkflowTable(graphs, core.RUNNING)
		}
	},
}

var listSuccessfulWorkflowsCmd = &cobra.Command{
	Use:   "pss",
	Short: "List all successful workflows",
	Long:  "List all successful workflows",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		graphs, err := client.GetSuccessfulProcessGraphs(ColonyName, Count, PrvKey)
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

			printWorkflowTable(graphs, core.SUCCESS)
		}
	},
}

var listFailedWorkflowsCmd = &cobra.Command{
	Use:   "psf",
	Short: "List all failed workflows",
	Long:  "List all failed workflows",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		graphs, err := client.GetFailedProcessGraphs(ColonyName, Count, PrvKey)
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

			printWorkflowTable(graphs, core.FAILED)
		}

	},
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
