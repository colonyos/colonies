package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/internal/crypto"
	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/kataras/tablewriter"
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
	rootCmd.AddCommand(processCmd)

	processCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	processCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	listWaitingProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listWaitingProcessesCmd.Flags().StringVarP(&TargetExecutorType, "type", "", "", "Only show processes targeting this executor type")
	listWaitingProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listWaitingProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	listWaitingProcessesCmd.Flags().BoolVarP(&ShowIDs, "ids", "i", false, "Show IDs")

	listRunningProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listRunningProcessesCmd.Flags().StringVarP(&TargetExecutorType, "type", "", "", "Only show processes targeting this executor type")
	listRunningProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listRunningProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	listRunningProcessesCmd.Flags().BoolVarP(&ShowIDs, "ids", "i", false, "Show IDs")

	listSuccessfulProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listSuccessfulProcessesCmd.Flags().StringVarP(&TargetExecutorType, "type", "", "", "Only show processes targeting this executor type")
	listSuccessfulProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listSuccessfulProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	listSuccessfulProcessesCmd.Flags().BoolVarP(&ShowIDs, "ids", "i", false, "Show IDs")

	listFailedProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	listFailedProcessesCmd.Flags().StringVarP(&ExecutorType, "type", "", "", "Only show processes targeting this executor type")
	listFailedProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listFailedProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	listFailedProcessesCmd.Flags().BoolVarP(&ShowIDs, "ids", "i", false, "Show IDs")

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
}

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Manage processes",
	Long:  "Manage processes",
}

func wait(client *client.ColoniesClient, process *core.Process) {
	for {
		successSubscription, err := client.SubscribeProcess(process.ID,
			process.FunctionSpec.Conditions.ExecutorType,
			core.SUCCESS,
			100,
			PrvKey)
		CheckError(err)

		failedSubscription, err := client.SubscribeProcess(process.ID,
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

var listWaitingProcessesCmd = &cobra.Command{
	Use:   "psw",
	Short: "List all waiting processes",
	Long:  "List all waiting processes",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		processes, err := client.GetWaitingProcesses(ColonyName, TargetExecutorType, Count, PrvKey)
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

			var data [][]string
			for _, process := range processes {
				args, kwArgs := parseArgs(process)
				if ShowIDs {
					data = append(data, []string{process.ID, process.FunctionSpec.FuncName, args, kwArgs, process.SubmissionTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType, process.InitiatorName})
				} else {
					data = append(data, []string{process.FunctionSpec.FuncName, args, kwArgs, process.SubmissionTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType, process.InitiatorName})
				}
			}
			table := tablewriter.NewWriter(os.Stdout)
			if ShowIDs {
				table.SetHeader([]string{"ID", "Func", "Args", "KwArgs", "Submission Time", "Executor Type", "Initiator Name"})
			} else {
				table.SetHeader([]string{"Func", "Args", "KwArgs", "Submission Time", "Executor Type", "Initiator Name"})
			}
			for _, v := range data {
				table.Append(v)
			}
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
		}

	},
}

var listRunningProcessesCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all running processes",
	Long:  "List all running processes",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		processes, err := client.GetRunningProcesses(ColonyName, TargetExecutorType, Count, PrvKey)
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

			var data [][]string
			for _, process := range processes {
				args, kwArgs := parseArgs(process)
				if ShowIDs {
					data = append(data, []string{process.ID, process.FunctionSpec.FuncName, args, kwArgs, process.SubmissionTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType, process.InitiatorName})
				} else {
					data = append(data, []string{process.FunctionSpec.FuncName, args, kwArgs, process.SubmissionTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType, process.InitiatorName})
				}
			}
			table := tablewriter.NewWriter(os.Stdout)
			if ShowIDs {
				table.SetHeader([]string{"ID", "FuncName", "Args", "KwArgs", "Start time", "Executor Type", "Initiator Name"})
			} else {
				table.SetHeader([]string{"FuncName", "Args", "KwArgs", "Start time", "Executor Type", "Initiator Name"})
			}
			for _, v := range data {
				table.Append(v)
			}
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
		}
	},
}

var listSuccessfulProcessesCmd = &cobra.Command{
	Use:   "pss",
	Short: "List all successful processes",
	Long:  "List all successful processes",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		processes, err := client.GetSuccessfulProcesses(ColonyName, TargetExecutorType, Count, PrvKey)
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

			var data [][]string
			for _, process := range processes {
				args, kwArgs := parseArgs(process)
				if ShowIDs {
					data = append(data, []string{process.ID, process.FunctionSpec.FuncName, args, kwArgs, process.SubmissionTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType, process.InitiatorName})
				} else {
					data = append(data, []string{process.FunctionSpec.FuncName, args, kwArgs, process.SubmissionTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType, process.InitiatorName})
				}
			}
			table := tablewriter.NewWriter(os.Stdout)
			if ShowIDs {
				table.SetHeader([]string{"ID", "FuncName", "Args", "KwArgs", "End time", "Executor Type", "Initiator Name"})
			} else {
				table.SetHeader([]string{"FuncName", "Args", "KwArgs", "End time", "Executor Type", "Initiator Name"})
			}
			for _, v := range data {
				table.Append(v)
			}
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
		}
	},
}

var listFailedProcessesCmd = &cobra.Command{
	Use:   "psf",
	Short: "List all failed processes",
	Long:  "List all failed processes",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		processes, err := client.GetFailedProcesses(ColonyName, TargetExecutorType, Count, PrvKey)
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

			var data [][]string
			for _, process := range processes {
				args, kwArgs := parseArgs(process)
				if ShowIDs {
					data = append(data, []string{process.ID, process.FunctionSpec.FuncName, args, kwArgs, process.SubmissionTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType, process.InitiatorName})
				} else {
					data = append(data, []string{process.FunctionSpec.FuncName, args, kwArgs, process.SubmissionTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType, process.InitiatorName})
				}
			}
			table := tablewriter.NewWriter(os.Stdout)
			if ShowIDs {
				table.SetHeader([]string{"ID", "FuncName", "Args", "KwArgs", "End time", "Executor Type", "Initiator"})
			} else {
				table.SetHeader([]string{"FuncName", "Args", "KwArgs", "End time", "Executor Type", "Initiator"})
			}
			for _, v := range data {
				table.Append(v)
			}
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
		}
	},
}

func printFunctionSpec(funcSpec *core.FunctionSpec) {
	executorIDs := ""
	for _, executorID := range funcSpec.Conditions.ExecutorIDs {
		executorIDs += executorID + "\n"
	}
	executorIDs = strings.TrimSuffix(executorIDs, "\n")
	if executorIDs == "" {
		executorIDs = "None"
	}

	procFunc := funcSpec.FuncName
	if procFunc == "" {
		procFunc = "None"
	}

	procArgs := ""
	for _, procArg := range IfArr2StringArr(funcSpec.Args) {
		procArgs += procArg + " "
	}
	if procArgs == "" {
		procArgs = "None"
	}

	if len(procArgs) > MaxArgInfoLength {
		procArgs = procArgs[0:MaxArgInfoLength] + "..."
	}

	procKwArgs := ""
	for k, procKwArg := range IfMap2StringMap(funcSpec.KwArgs) {
		procKwArgs += k + ":" + procKwArg + " "
	}
	if procKwArgs == "" {
		procKwArgs = "None"
	}

	if len(procKwArgs) > MaxArgInfoLength {
		procKwArgs = procKwArgs[0:MaxArgInfoLength] + "..."
	}

	specData := [][]string{
		[]string{"Func", procFunc},
		[]string{"Args", procArgs},
		[]string{"KwArgs", procKwArgs},
		[]string{"MaxWaitTime", strconv.Itoa(funcSpec.MaxWaitTime)},
		[]string{"MaxExecTime", strconv.Itoa(funcSpec.MaxExecTime)},
		[]string{"MaxRetries", strconv.Itoa(funcSpec.MaxRetries)},
		[]string{"Priority", strconv.Itoa(funcSpec.Priority)},
	}
	specTable := tablewriter.NewWriter(os.Stdout)
	for _, v := range specData {
		specTable.Append(v)
	}
	specTable.SetAlignment(tablewriter.ALIGN_LEFT)
	specTable.Render()

	fmt.Println()
	fmt.Println("Conditions:")

	dep := ""
	for _, s := range funcSpec.Conditions.Dependencies {
		dep += s + " "
	}
	if len(dep) > 0 {
		dep = dep[:len(dep)-1]
	}

	condData := [][]string{
		[]string{"ColonyName", funcSpec.Conditions.ColonyName},
		[]string{"ExecutorIDs", executorIDs},
		[]string{"ExecutorType", funcSpec.Conditions.ExecutorType},
		[]string{"Dependencies", dep},
		[]string{"Nodes", strconv.Itoa(funcSpec.Conditions.Nodes)},
		[]string{"CPU", funcSpec.Conditions.CPU},
		[]string{"Memmory", funcSpec.Conditions.Memory},
		[]string{"Processes", strconv.Itoa(funcSpec.Conditions.Processes)},
		[]string{"ProcessesPerNode", strconv.Itoa(funcSpec.Conditions.ProcessesPerNode)},
		[]string{"Storage", funcSpec.Conditions.Storage},
		[]string{"Walltime", strconv.Itoa(int(funcSpec.Conditions.WallTime))},
		[]string{"GPU", funcSpec.Conditions.GPU.Name},
		[]string{"GPUs", strconv.Itoa(int(funcSpec.Conditions.GPU.Count))},
		[]string{"GPUMemory", funcSpec.Conditions.GPU.Memory},
	}
	condTable := tablewriter.NewWriter(os.Stdout)
	for _, v := range condData {
		condTable.Append(v)
	}
	condTable.SetAlignment(tablewriter.ALIGN_LEFT)
	condTable.Render()
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

		assignedExecutorID := "None"
		if process.AssignedExecutorID != "" {
			assignedExecutorID = process.AssignedExecutorID
		}

		isAssigned := "False"
		if process.IsAssigned {
			isAssigned = "True"
		}

		if PrintOutput {
			fmt.Println(StrArr2Str(IfArr2StringArr(process.Output)))
			os.Exit(0)
		}

		input := StrArr2Str(IfArr2StringArr(process.Input))
		if len(input) > MaxArgInfoLength {
			input = input[0:MaxArgInfoLength] + "..."
		}

		output := StrArr2Str(IfArr2StringArr(process.Output))
		if len(output) > MaxArgInfoLength {
			output = output[0:MaxArgInfoLength] + "..."
		}

		fmt.Println("Process:")
		processData := [][]string{
			[]string{"ID", process.ID},
			[]string{"IsAssigned", isAssigned},
			[]string{"InitiatorID", process.InitiatorID},
			[]string{"InitiatorName", process.InitiatorName},
			[]string{"AssignedExecutorID", assignedExecutorID},
			[]string{"State", State2String(process.State)},
			[]string{"PriorityTime", strconv.FormatInt(process.PriorityTime, 10)},
			[]string{"SubmissionTime", process.SubmissionTime.Format(TimeLayout)},
			[]string{"StartTime", process.StartTime.Format(TimeLayout)},
			[]string{"EndTime", process.EndTime.Format(TimeLayout)},
			[]string{"WaitDeadline", process.WaitDeadline.Format(TimeLayout)},
			[]string{"ExecDeadline", process.ExecDeadline.Format(TimeLayout)},
			[]string{"WaitingTime", process.WaitingTime().String()},
			[]string{"ProcessingTime", process.ProcessingTime().String()},
			[]string{"Retries", strconv.Itoa(process.Retries)},
			[]string{"Input", input},
			[]string{"Output", output},
			[]string{"Errors", StrArr2Str(process.Errors)},
		}
		processTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range processData {
			processTable.Append(v)
		}
		processTable.SetAlignment(tablewriter.ALIGN_LEFT)
		processTable.Render()

		fmt.Println()
		fmt.Println("FunctionSpec:")
		printFunctionSpec(&process.FunctionSpec)

		fmt.Println()
		fmt.Println("Attributes:")
		if len(process.Attributes) > 0 {

			var attributeData [][]string
			for _, attribute := range process.Attributes {
				var attributeType string
				switch attribute.AttributeType {
				case core.IN:
					attributeType = "In"
				case core.OUT:
					attributeType = "Out"
				case core.ERR:
					attributeType = "Err"
				case core.ENV:
					attributeType = "Env"
				default:
					attributeType = "Unknown"
				}
				var key string
				if len(attribute.Key) > MaxAttributeLength {
					key = attribute.Key[0:MaxAttributeLength] + "..."
				} else {
					key = attribute.Key
				}

				var value string
				if len(attribute.Value) > MaxAttributeLength {
					value = attribute.Value[0:MaxAttributeLength] + "..."
				} else {
					value = attribute.Value
				}
				attributeData = append(attributeData, []string{attribute.ID, key, value, attributeType})
			}

			attributeTable := tablewriter.NewWriter(os.Stdout)
			attributeTable.SetHeader([]string{"ID", "Key", "Value", "Type"})
			attributeTable.SetAlignment(tablewriter.ALIGN_LEFT)
			for _, v := range attributeData {
				attributeTable.Append(v)
			}
			attributeTable.Render()
		} else {
			fmt.Println("No attributes found")
		}
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

		fmt.Print("WARNING!!! Are you sure you want to remove " + state + " processes from Colony <" + ColonyName + ">. This operation cannot be undone! (YES,no): ")

		var err error
		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')
		if reply == "YES\n" {
			if state == "all" {
				err = client.RemoveAllProcesses(ColonyName, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all processes in Colony <" + ColonyName + ">")
			} else if Waiting {
				err = client.RemoveAllProcessesWithState(ColonyName, core.WAITING, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all waiting processes in Colony <" + ColonyName + ">")
			} else if Successful {
				err = client.RemoveAllProcessesWithState(ColonyName, core.SUCCESS, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all successful processes in Colony <" + ColonyName + ">")
			} else if Failed {
				err = client.RemoveAllProcessesWithState(ColonyName, core.FAILED, ColonyPrvKey)
				CheckError(err)
				log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("Removing all failed processes in Colony <" + ColonyName + ">")
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

		log.WithFields(log.Fields{"ProcessId": process.ID}).Info("Process closed as Successful")
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

		log.WithFields(log.Fields{"ProcessId": process.ID}).Info("Process closed as Failed")
	},
}
