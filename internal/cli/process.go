package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/colonyos/colonies/internal/crypto"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	processCmd.AddCommand(listWaitingProcessesCmd)
	processCmd.AddCommand(listRunningProcessesCmd)
	processCmd.AddCommand(listSuccessfulProcessesCmd)
	processCmd.AddCommand(listFailedProcessesCmd)
	processCmd.AddCommand(getProcessCmd)
	processCmd.AddCommand(removeProcessCmd)
	processCmd.AddCommand(removeAllProcessesCmd)
	processCmd.AddCommand(assignProcessCmd)
	processCmd.AddCommand(closeSuccessfulCmd)
	processCmd.AddCommand(closeFailedCmd)
	processCmd.AddCommand(pauseAssignmentsCmd)
	processCmd.AddCommand(resumeAssignmentsCmd)
	processCmd.AddCommand(statusAssignmentsCmd)
	rootCmd.AddCommand(processCmd)

	processCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	processCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	listWaitingProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listWaitingProcessesCmd.Flags().IntVarP(&Count, "count", "", DefaultCount, "Number of processes to list")
	listWaitingProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	listWaitingProcessesCmd.Flags().BoolVarP(&ShowIDs, "ids", "i", false, "Show IDs")
	listWaitingProcessesCmd.Flags().StringVarP(&TargetExecutorType, "executortype", "", "", "Filter by executor type")
	listWaitingProcessesCmd.Flags().StringVarP(&Label, "label", "", "", "Filter by label")
	listWaitingProcessesCmd.Flags().StringVarP(&Initiator, "initiator", "", "", "Filter by initiator")

	listRunningProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listRunningProcessesCmd.Flags().IntVarP(&Count, "count", "", DefaultCount, "Number of processes to list")
	listRunningProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	listRunningProcessesCmd.Flags().BoolVarP(&ShowIDs, "ids", "i", false, "Show IDs")
	listRunningProcessesCmd.Flags().StringVarP(&TargetExecutorType, "executortype", "", "", "Filter by executor type")
	listRunningProcessesCmd.Flags().StringVarP(&Label, "label", "", "", "Filter by label")
	listRunningProcessesCmd.Flags().StringVarP(&Initiator, "initiator", "", "", "Filter by initiator")

	listSuccessfulProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listSuccessfulProcessesCmd.Flags().IntVarP(&Count, "count", "", DefaultCount, "Number of processes to list")
	listSuccessfulProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	listSuccessfulProcessesCmd.Flags().BoolVarP(&ShowIDs, "ids", "i", false, "Show IDs")
	listSuccessfulProcessesCmd.Flags().StringVarP(&TargetExecutorType, "executortype", "", "", "Filter by executor type")
	listSuccessfulProcessesCmd.Flags().StringVarP(&Label, "label", "", "", "Filter by label")
	listSuccessfulProcessesCmd.Flags().StringVarP(&Initiator, "initiator", "", "", "Filter by initiator")

	listFailedProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listFailedProcessesCmd.Flags().IntVarP(&Count, "count", "", DefaultCount, "Number of processes to list")
	listFailedProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	listFailedProcessesCmd.Flags().BoolVarP(&ShowIDs, "ids", "i", false, "Show IDs")
	listFailedProcessesCmd.Flags().StringVarP(&TargetExecutorType, "executortype", "", "", "Filter by executor type")
	listFailedProcessesCmd.Flags().StringVarP(&Label, "label", "", "", "Filter by label")
	listFailedProcessesCmd.Flags().StringVarP(&Initiator, "initiator", "", "", "Filter by initiator")

	getProcessCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	getProcessCmd.MarkFlagRequired("processid")
	getProcessCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	getProcessCmd.Flags().BoolVarP(&PrintOutput, "out", "", false, "Print process output")

	removeProcessCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	removeProcessCmd.MarkFlagRequired("processid")

	removeAllProcessesCmd.Flags().BoolVarP(&Waiting, "waiting", "", false, "Remove all waiting processes")
	removeAllProcessesCmd.Flags().BoolVarP(&Successful, "successful", "", false, "Remove all successful processes")
	removeAllProcessesCmd.Flags().BoolVarP(&Failed, "failed", "", false, "Remove all failed processes")

	assignProcessCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	assignProcessCmd.Flags().IntVarP(&Timeout, "timeout", "", 100, "Max time to wait for a process assignment")

	closeSuccessfulCmd.Flags().StringSliceVarP(&Output, "out", "", make([]string, 0), "Output")
	closeSuccessfulCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	closeSuccessfulCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	closeSuccessfulCmd.MarkFlagRequired("processid")

	closeFailedCmd.Flags().StringSliceVarP(&Errors, "errors", "", make([]string, 0), "Errors")
	closeFailedCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	closeFailedCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	closeFailedCmd.MarkFlagRequired("processid")

	pauseAssignmentsCmd.Flags().StringVarP(&ColonyName, "colonyname", "", "", "Colony name")
	pauseAssignmentsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	resumeAssignmentsCmd.Flags().StringVarP(&ColonyName, "colonyname", "", "", "Colony name")
	resumeAssignmentsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	statusAssignmentsCmd.Flags().StringVarP(&ColonyName, "colonyname", "", "", "Colony name")
	statusAssignmentsCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
}

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Manage processes",
	Long:  "Manage processes",
}

func wait(client *client.ColoniesClient, process *core.Process) {
	for {
		successSubscription, err := client.SubscribeProcess(
			ColonyName,
			process.ID,
			process.FunctionSpec.Conditions.ExecutorType,
			core.SUCCESS,
			100,
			PrvKey)
		CheckError(err)

		failedSubscription, err := client.SubscribeProcess(
			ColonyName,
			process.ID,
			process.FunctionSpec.Conditions.ExecutorType,
			core.FAILED,
			100,
			PrvKey)
		CheckError(err)

		select {
		case <-successSubscription.ProcessChan:
			return
		case err := <-successSubscription.ErrChan:
			CheckError(err)
		case <-failedSubscription.ProcessChan:
			return
		case err := <-failedSubscription.ErrChan:
			CheckError(err)
		}
	}
}

var assignProcessCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a process to a executor",
	Long:  "Assign a process to a executor",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		executorID, err := crypto.GenerateID(PrvKey)
		CheckError(err)

		process, err := client.Assign(ColonyName, Timeout, "", "", PrvKey)
		if err != nil {
			log.Warning(err)
		} else {
			log.WithFields(log.Fields{"ProcessId": process.ID, "ExecutorId": executorID}).Info("Assigned process to executor")
		}
	},
}

func checkFilterArgs() {
	counter := 0

	if TargetExecutorType != "" {
		counter++
	}

	if Label != "" {
		counter++
	}

	if Initiator != "" {
		counter++
	}

	if counter > 1 {
		CheckError(errors.New("Invalid filter arguments, select --executortype, --label or --initiator"))
	}
}

var listWaitingProcessesCmd = &cobra.Command{
	Use:   "psw",
	Short: "List all waiting processes",
	Long:  "List all waiting processes",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		checkFilterArgs()

		processes, err := client.GetWaitingProcesses(ColonyName, TargetExecutorType, Label, Initiator, Count, PrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("No waiting processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			printProcessesTable(processes, core.WAITING)
		}
	},
}

var listRunningProcessesCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all running processes",
	Long:  "List all running processes",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		checkFilterArgs()

		processes, err := client.GetRunningProcesses(ColonyName, TargetExecutorType, Label, Initiator, Count, PrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("No running processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			printProcessesTable(processes, core.RUNNING)
		}
	},
}

var listSuccessfulProcessesCmd = &cobra.Command{
	Use:   "pss",
	Short: "List all successful processes",
	Long:  "List all successful processes",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		checkFilterArgs()

		processes, err := client.GetSuccessfulProcesses(ColonyName, TargetExecutorType, Label, Initiator, Count, PrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("No successful processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			printProcessesTable(processes, core.WAITING)
		}
	},
}

var listFailedProcessesCmd = &cobra.Command{
	Use:   "psf",
	Short: "List all failed processes",
	Long:  "List all failed processes",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		checkFilterArgs()

		processes, err := client.GetFailedProcesses(ColonyName, TargetExecutorType, Label, Initiator, Count, PrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("No failed processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			printProcessesTable(processes, core.FAILED)
		}
	},
}

var getProcessCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a process",
	Long:  "Get info about a process",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		process, err := client.GetProcess(ProcessID, PrvKey)
		if err != nil {
			log.WithFields(log.Fields{"ProcessId": ProcessID, "Error": err}).Info("Process not found")
			os.Exit(-1)
		}

		if JSON {
			fmt.Println(process.ToJSON())
			os.Exit(0)
		}

		if PrintOutput {
			fmt.Println(StrArr2Str(IfArr2StringArr(process.Output)))
			os.Exit(0)
		}

		printProcessTable(process)
		printFunctionSpecTable(&process.FunctionSpec)
		printConditionsTable(&process.FunctionSpec)
		printAttributesTable(process)
	},
}

var removeProcessCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a process",
	Long:  "Remove a process",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.RemoveProcess(ProcessID, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessId": ProcessID}).Info("Process removed")
	},
}

var removeAllProcessesCmd = &cobra.Command{
	Use:   "removeall",
	Short: "Remove all processes",
	Long:  "Remove all processes",
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

		fmt.Print("WARNING!!! Are you sure you want to remove " + state + " processes in colony <" + ColonyName + ">. This operation cannot be undone! (YES,no): ")

		var err error
		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')
		if reply == "YES\n" {
			if state == "all" {
				err = client.RemoveAllProcesses(ColonyName, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all processes in colony <" + ColonyName + ">")
			} else if Waiting {
				err = client.RemoveAllProcessesWithState(ColonyName, core.WAITING, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all waiting processes in colony <" + ColonyName + ">")
			} else if Successful {
				err = client.RemoveAllProcessesWithState(ColonyName, core.SUCCESS, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all successful processes in colony <" + ColonyName + ">")
			} else if Failed {
				err = client.RemoveAllProcessesWithState(ColonyName, core.FAILED, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all failed processes in colony <" + ColonyName + ">")
			}
		} else {
			log.Info("Aborting ...")
		}
	},
}

var closeSuccessfulCmd = &cobra.Command{
	Use:   "close",
	Short: "Close a process as successful",
	Long:  "Close a process as successful",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		process, err := client.GetProcess(ProcessID, PrvKey)
		CheckError(err)

		if len(Output) > 0 {
			outputIf := make([]interface{}, len(Output))
			for k, v := range Output {
				outputIf[k] = v
			}
			err = client.CloseWithOutput(process.ID, outputIf, PrvKey)
			CheckError(err)
		} else {
			err = client.Close(process.ID, PrvKey)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ProcessId": process.ID}).Info("Process closed as successful")
	},
}

var closeFailedCmd = &cobra.Command{
	Use:   "fail",
	Short: "Close a process as failed",
	Long:  "Close a process as failed",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		process, err := client.GetProcess(ProcessID, PrvKey)
		CheckError(err)

		if len(Errors) == 0 {
			Errors = []string{"No errors specified"}
		}

		err = client.Fail(process.ID, Errors, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessId": process.ID}).Info("Process closed as failed")
	},
}

var pauseAssignmentsCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause process assignments for a colony",
	Long:  "Pause all process assignments for the specified colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()
		err := client.PauseColonyAssignments(ColonyName, PrvKey)
		CheckError(err)
		log.WithFields(log.Fields{"Colony": ColonyName}).Info("Colony process assignments have been paused")
	},
}

var resumeAssignmentsCmd = &cobra.Command{
	Use:   "resume", 
	Short: "Resume process assignments for a colony",
	Long:  "Resume all process assignments for the specified colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()
		err := client.ResumeColonyAssignments(ColonyName, PrvKey)
		CheckError(err)
		log.WithFields(log.Fields{"Colony": ColonyName}).Info("Colony process assignments have been resumed")
	},
}

var statusAssignmentsCmd = &cobra.Command{
	Use:   "status",
	Short: "Check if process assignments are paused for a colony",
	Long:  "Check the pause status of process assignments for the specified colony",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()
		isPaused, err := client.AreColonyAssignmentsPaused(ColonyName, PrvKey)
		CheckError(err)
		if isPaused {
			log.WithFields(log.Fields{"Colony": ColonyName}).Info("Colony process assignments are PAUSED")
		} else {
			log.WithFields(log.Fields{"Colony": ColonyName}).Info("Colony process assignments are ACTIVE")
		}
	},
}
