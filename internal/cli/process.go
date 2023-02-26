package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
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
	processCmd.AddCommand(deleteProcessCmd)
	processCmd.AddCommand(deleteAllProcessesCmd)
	processCmd.AddCommand(assignProcessCmd)
	processCmd.AddCommand(closeSuccessfulCmd)
	processCmd.AddCommand(closeFailedCmd)
	rootCmd.AddCommand(processCmd)

	processCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	processCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	listWaitingProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listWaitingProcessesCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	listWaitingProcessesCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	listWaitingProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listWaitingProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	listRunningProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listRunningProcessesCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	listRunningProcessesCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	listRunningProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listRunningProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	listSuccessfulProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listSuccessfulProcessesCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	listSuccessfulProcessesCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	listSuccessfulProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listSuccessfulProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	listFailedProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listFailedProcessesCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	listFailedProcessesCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	listFailedProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listFailedProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	getProcessCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	getProcessCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	getProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	getProcessCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	getProcessCmd.MarkFlagRequired("processid")
	getProcessCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	getProcessCmd.Flags().BoolVarP(&PrintOutput, "out", "", false, "Print process output")

	deleteProcessCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	deleteProcessCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	deleteProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	deleteProcessCmd.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	deleteProcessCmd.MarkFlagRequired("processid")

	deleteAllProcessesCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	deleteAllProcessesCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	deleteAllProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")

	assignProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	assignProcessCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	assignProcessCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	assignProcessCmd.Flags().IntVarP(&Timeout, "timeout", "", 100, "Max time to wait for a process assignment")
	assignProcessCmd.Flags().BoolVarP(&Latest, "latest", "", false, "Try to assign the latest process in the queue")

	closeSuccessfulCmd.Flags().StringSliceVarP(&Output, "out", "", make([]string, 0), "Output")
	closeSuccessfulCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	closeSuccessfulCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	closeSuccessfulCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	closeSuccessfulCmd.MarkFlagRequired("processid")

	closeFailedCmd.Flags().StringSliceVarP(&Errors, "errors", "", make([]string, 0), "Errors")
	closeFailedCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	closeFailedCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
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
		subscription, err := client.SubscribeProcess(process.ID,
			process.FunctionSpec.Conditions.ExecutorType,
			core.SUCCESS,
			100,
			ExecutorPrvKey)
		CheckError(err)

		select {
		case process := <-subscription.ProcessChan:
			for _, attribute := range process.Attributes {
				if attribute.Key == "output" {
					fmt.Print(attribute.Value)
				}
			}
			os.Exit(0)
		case err := <-subscription.ErrChan:
			CheckError(err)
		}
	}

}

var assignProcessCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a process to a executor",
	Long:  "Assign a process to a executor",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		if Latest {
			process, err := client.AssignLatestProcess(ColonyID, Timeout, ExecutorPrvKey)
			if err != nil {
				log.Warning(err)
			} else {
				log.WithFields(log.Fields{"ProcessID": process.ID, "ExecutorID": ExecutorID}).Info("Assigned process to executor (latest)")
			}
		} else {
			process, err := client.Assign(ColonyID, Timeout, ExecutorPrvKey)
			if err != nil {
				log.Warning(err)
			} else {
				log.WithFields(log.Fields{"ProcessID": process.ID, "ExecutorID": ExecutorID}).Info("Assigned process to executor (oldest)")
			}
		}

	},
}

var listWaitingProcessesCmd = &cobra.Command{
	Use:   "psw",
	Short: "List all waiting processes",
	Long:  "List all waiting processes",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		processes, err := client.GetWaitingProcesses(ColonyID, Count, ExecutorPrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.WithFields(log.Fields{"ColonyId": ColonyID}).Info("No waiting processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.FunctionSpec.FuncName, StrArr2Str(process.FunctionSpec.Args), process.SubmissionTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Func", "Args", "Submission Time", "Executor Type"})
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		processes, err := client.GetRunningProcesses(ColonyID, Count, ExecutorPrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.WithFields(log.Fields{"ColonyId": ColonyID}).Info("No running processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.FunctionSpec.FuncName, StrArr2Str(process.FunctionSpec.Args), process.StartTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "FuncName", "Args", "Start time", "Executor Type"})
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		processes, err := client.GetSuccessfulProcesses(ColonyID, Count, ExecutorPrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.WithFields(log.Fields{"ColonyId": ColonyID}).Info("No successful processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.FunctionSpec.FuncName, StrArr2Str(process.FunctionSpec.Args), process.EndTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "FuncName", "Args", "End time", "Executor Type"})
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		processes, err := client.GetFailedProcesses(ColonyID, Count, ExecutorPrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.WithFields(log.Fields{"ColonyId": ColonyID}).Info("No failed processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.FunctionSpec.FuncName, StrArr2Str(process.FunctionSpec.Args), process.EndTime.Format(TimeLayout), process.FunctionSpec.Conditions.ExecutorType})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "FuncName", "Args", "End time", "Executor Type"})
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
	for _, procArg := range funcSpec.Args {
		procArgs += procArg + " "
	}
	if procArgs == "" {
		procArgs = "None"
	}

	specData := [][]string{
		[]string{"Func", procFunc},
		[]string{"Args", procArgs},
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
		[]string{"ColonyID", funcSpec.Conditions.ColonyID},
		[]string{"ExecutorIDs", executorIDs},
		[]string{"ExecutorType", funcSpec.Conditions.ExecutorType},
		[]string{"Dependencies", dep},
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}
		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		process, err := client.GetProcess(ProcessID, ExecutorPrvKey)
		if err != nil {
			log.WithFields(log.Fields{"ProcessID": ProcessID, "Error": err}).Info("Process not found")
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
			fmt.Println(StrArr2Str(process.Output))
			os.Exit(0)
		}

		fmt.Println("Process:")

		processData := [][]string{
			[]string{"ID", process.ID},
			[]string{"IsAssigned", isAssigned},
			[]string{"AssignedExecutorID", assignedExecutorID},
			[]string{"State", State2String(process.State)},
			[]string{"Priority", strconv.Itoa(process.FunctionSpec.Priority)},
			[]string{"SubmissionTime", process.SubmissionTime.Format(TimeLayout)},
			[]string{"StartTime", process.StartTime.Format(TimeLayout)},
			[]string{"EndTime", process.EndTime.Format(TimeLayout)},
			[]string{"WaitDeadline", process.WaitDeadline.Format(TimeLayout)},
			[]string{"ExecDeadline", process.ExecDeadline.Format(TimeLayout)},
			[]string{"WaitingTime", process.WaitingTime().String()},
			[]string{"ProcessingTime", process.ProcessingTime().String()},
			[]string{"Retries", strconv.Itoa(process.Retries)},
			[]string{"Errors", StrArr2Str(process.Errors)},
			[]string{"Output", StrArr2Str(process.Output)},
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

var deleteProcessCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a process",
	Long:  "Delete a process",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}
		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		err = client.DeleteProcess(ProcessID, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessID": ProcessID}).Info("Process deleted")
	},
}

var deleteAllProcessesCmd = &cobra.Command{
	Use:   "deleteall",
	Short: "Delete all processes",
	Long:  "Delete all processes",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		fmt.Print("WARNING!!! Are you sure you want to delete all process in the Colony This operation cannot be undone! (YES,no): ")

		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')
		if reply == "YES\n" {
			log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
			client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

			err = client.DeleteAllProcesses(ColonyID, ColonyPrvKey)
			CheckError(err)

			log.WithFields(log.Fields{"ColonyID": ColonyID}).Info("Deleting all processes in Colony")
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		process, err := client.GetProcess(ProcessID, ExecutorPrvKey)
		CheckError(err)

		if len(Output) > 0 {
			err = client.CloseWithOutput(process.ID, Output, ExecutorPrvKey)
			CheckError(err)
		} else {
			err = client.Close(process.ID, ExecutorPrvKey)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Process closed as Successful")
	},
}

var closeFailedCmd = &cobra.Command{
	Use:   "fail",
	Short: "Close a process as failed",
	Long:  "Close a process as failed",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		process, err := client.GetProcess(ProcessID, ExecutorPrvKey)
		CheckError(err)

		if len(Errors) == 0 {
			Errors = []string{"No errors specified"}
		}

		err = client.Fail(process.ID, Errors, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessID": process.ID}).Info("Process closed as Failed")
	},
}
