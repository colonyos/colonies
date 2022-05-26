package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	processCmd.AddCommand(submitProcessCmd)
	processCmd.AddCommand(listWaitingProcessesCmd)
	processCmd.AddCommand(listRunningProcessesCmd)
	processCmd.AddCommand(listSuccessfulProcessesCmd)
	processCmd.AddCommand(listFailedProcessesCmd)
	processCmd.AddCommand(getProcessCmd)
	processCmd.AddCommand(deleteProcessCmd)
	processCmd.AddCommand(deleteAllProcessesCmd)
	processCmd.AddCommand(assignProcessCmd)
	processCmd.AddCommand(closeSuccessful)
	processCmd.AddCommand(closeFailed)
	rootCmd.AddCommand(processCmd)

	processCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	processCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 50080, "Server HTTP port")

	submitProcessCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	submitProcessCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	submitProcessCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony process")
	submitProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	submitProcessCmd.Flags().BoolVarP(&Wait, "wait", "", false, "Colony Id")
	submitProcessCmd.MarkFlagRequired("spec")

	listWaitingProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listWaitingProcessesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listWaitingProcessesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listWaitingProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listWaitingProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	listRunningProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listRunningProcessesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listRunningProcessesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listRunningProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listRunningProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	listSuccessfulProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listSuccessfulProcessesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listSuccessfulProcessesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listSuccessfulProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listSuccessfulProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	listFailedProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listFailedProcessesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listFailedProcessesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listFailedProcessesCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of processes to list")
	listFailedProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	getProcessCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	getProcessCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	getProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	getProcessCmd.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	getProcessCmd.MarkFlagRequired("processid")
	getProcessCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	getProcessCmd.Flags().BoolVarP(&Output, "output", "", false, "Get process output")

	deleteProcessCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	deleteProcessCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	deleteProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	deleteProcessCmd.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	deleteProcessCmd.MarkFlagRequired("processid")

	deleteAllProcessesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	deleteAllProcessesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	deleteAllProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")

	assignProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	assignProcessCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	assignProcessCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")

	closeSuccessful.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	closeSuccessful.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	closeSuccessful.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	closeSuccessful.MarkFlagRequired("processid")

	closeFailed.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	closeFailed.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	closeFailed.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	closeFailed.MarkFlagRequired("processid")
}

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Manage Colony Processes",
	Long:  "Manage Colony Processes",
}

var submitProcessCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a Process to a Colony",
	Long:  "Submit a Process to a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		processSpec, err := core.ConvertJSONToProcessSpec(string(jsonSpecBytes))
		CheckError(err)

		if processSpec.Conditions.ColonyID == "" {
			if ColonyID == "" {
				ColonyID = os.Getenv("COLONIES_COLONYID")
			}
			if ColonyID == "" {
				CheckError(errors.New("Unknown Colony Id, please set COLONYID env variable or specify ColonyID in JSON file"))
			}

			processSpec.Conditions.ColonyID = ColonyID
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		addedProcess, err := client.SubmitProcessSpec(processSpec, RuntimePrvKey)
		CheckError(err)

		if Wait {
			for {
				processFromServer, err := client.GetProcess(addedProcess.ID, RuntimePrvKey)
				CheckError(err)

				if processFromServer.State == core.SUCCESS || processFromServer.State == core.FAILED {
					for _, attribute := range processFromServer.Attributes {
						if attribute.Key == "output" {
							fmt.Print(attribute.Value)
						}
					}
					os.Exit(0)
				} else {
					time.Sleep(500 * time.Millisecond)
				}
			}

		} else {
			log.WithFields(log.Fields{"processID": addedProcess.ID}).Info("Process submitted")
		}
	},
}

var assignProcessCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a Process to a Runtime",
	Long:  "Assign a Process to a Runtime",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		process, err := client.AssignProcess(ColonyID, RuntimePrvKey)
		if err != nil {
			log.Warning("No process was assigned")
		} else {
			log.WithFields(log.Fields{"processID": process.ID, "runtimeID": RuntimeID}).Info("Assigned process to runtime")
		}
	},
}

var listWaitingProcessesCmd = &cobra.Command{
	Use:   "psw",
	Short: "List all Waiting Processes",
	Long:  "List all Waiting Processes",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		processes, err := client.GetWaitingProcesses(ColonyID, Count, RuntimePrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.Warning("No waiting processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.ProcessSpec.Cmd, Args2String(process.ProcessSpec.Args), process.SubmissionTime.Format(TimeLayout), process.ProcessSpec.Conditions.RuntimeType})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Cmd", "Args", "Submission Time", "Runtime Type"})
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
	Short: "List all Running Processes",
	Long:  "List all Running Processes",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		processes, err := client.GetRunningProcesses(ColonyID, Count, RuntimePrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.Warning("No running processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.ProcessSpec.Cmd, Args2String(process.ProcessSpec.Args), process.StartTime.Format(TimeLayout), process.ProcessSpec.Conditions.RuntimeType})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Cmd", "Args", "Start time", "Runtime Type"})
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
	Short: "List all Successfull Processes",
	Long:  "List all Successfull Processes",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		processes, err := client.GetSuccessfulProcesses(ColonyID, Count, RuntimePrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.Warning("No successful processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.ProcessSpec.Cmd, Args2String(process.ProcessSpec.Args), process.EndTime.Format(TimeLayout), process.ProcessSpec.Conditions.RuntimeType})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Cmd", "Args", "End time", "Runtime Type"})
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
	Short: "List all Failed Processes",
	Long:  "List all Failed Processes",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		processes, err := client.GetFailedProcesses(ColonyID, Count, RuntimePrvKey)
		CheckError(err)

		if len(processes) == 0 {
			log.Warning("No failed processes found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.ProcessSpec.Cmd, Args2String(process.ProcessSpec.Args), process.EndTime.Format(TimeLayout), process.ProcessSpec.Conditions.RuntimeType})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Cmd", "Args", "End time", "Runtime Type"})
			for _, v := range data {
				table.Append(v)
			}
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.Render()
		}
	},
}

var getProcessCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a Process",
	Long:  "Get info about a Process",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		process, err := client.GetProcess(ProcessID, RuntimePrvKey)
		CheckError(err)
		if process == nil {
			log.Warning("Process not found")
			os.Exit(-1)
		}

		if JSON {
			fmt.Println(process.ToJSON())
			os.Exit(0)
		}

		assignedRuntimeID := "None"
		if process.AssignedRuntimeID != "" {
			assignedRuntimeID = process.AssignedRuntimeID
		}

		isAssigned := "False"
		if process.IsAssigned {
			isAssigned = "True"
		}

		var state string
		switch process.State {
		case core.WAITING:
			state = "Waiting"
		case core.RUNNING:
			state = "Running"
		case core.SUCCESS:
			state = "Successful"
		case core.FAILED:
			state = "Failed"
		default:
			state = "Unkown"
		}

		if Output {
			for _, attribute := range process.Attributes {
				if attribute.Key == "output" {
					fmt.Print(attribute.Value)
				}
			}
			os.Exit(0)
		}

		fmt.Println("Process:")

		processData := [][]string{
			[]string{"ID", process.ID},
			[]string{"IsAssigned", isAssigned},
			[]string{"AssignedRuntimeID", assignedRuntimeID},
			[]string{"State", state},
			[]string{"SubmissionTime", process.SubmissionTime.Format(TimeLayout)},
			[]string{"StartTime", process.StartTime.Format(TimeLayout)},
			[]string{"EndTime", process.EndTime.Format(TimeLayout)},
			[]string{"Deadline", process.Deadline.Format(TimeLayout)},
			[]string{"WaitingTime", process.WaitingTime().String()},
			[]string{"ProcessingTime", process.ProcessingTime().String()},
			[]string{"Retries", strconv.Itoa(process.Retries)},
		}
		processTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range processData {
			processTable.Append(v)
		}
		processTable.SetAlignment(tablewriter.ALIGN_LEFT)
		processTable.Render()

		runtimeIDs := ""
		for _, runtimeID := range process.ProcessSpec.Conditions.RuntimeIDs {
			runtimeIDs += runtimeID + "\n"
		}
		runtimeIDs = strings.TrimSuffix(runtimeIDs, "\n")
		if runtimeIDs == "" {
			runtimeIDs = "None"
		}

		image := process.ProcessSpec.Image
		if image == "" {
			image = "None"
		}

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

		volumes := ""
		for _, volume := range process.ProcessSpec.Volumes {
			volumes += volume + " "
		}
		if volumes == "" {
			volumes = "None"
		}

		ports := ""
		for _, port := range process.ProcessSpec.Ports {
			ports += port + " "
		}
		if ports == "" {
			ports = "None"
		}

		fmt.Println()
		fmt.Println("ProcessSpec:")

		specData := [][]string{
			[]string{"Image", image},
			[]string{"Cmd", procCmd},
			[]string{"Args", procArgs},
			[]string{"Volumes", volumes},
			[]string{"Ports", ports},
			[]string{"MaxExecTime", strconv.Itoa(process.ProcessSpec.MaxExecTime)},
			[]string{"MaxRetries", strconv.Itoa(process.ProcessSpec.MaxRetries)},
		}
		specTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range specData {
			specTable.Append(v)
		}
		specTable.SetAlignment(tablewriter.ALIGN_LEFT)
		specTable.Render()

		fmt.Println()
		fmt.Println("Conditions:")

		condData := [][]string{
			[]string{"ColonyID", process.ProcessSpec.Conditions.ColonyID},
			[]string{"RuntimeIDs", runtimeIDs},
			[]string{"RuntimeType", process.ProcessSpec.Conditions.RuntimeType},
			[]string{"Memory", strconv.Itoa(process.ProcessSpec.Conditions.Mem)},
			[]string{"CPU Cores", strconv.Itoa(process.ProcessSpec.Conditions.Cores)},
			[]string{"GPUs", strconv.Itoa(process.ProcessSpec.Conditions.GPUs)},
		}
		condTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range condData {
			condTable.Append(v)
		}
		condTable.SetAlignment(tablewriter.ALIGN_LEFT)
		condTable.Render()

		fmt.Println()
		fmt.Println("Attributes:")

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
	},
}

var deleteProcessCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a Process",
	Long:  "Delete a Process",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		err = client.DeleteProcess(ProcessID, RuntimePrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"processID": ProcessID}).Info("Process deleted")
	},
}

var deleteAllProcessesCmd = &cobra.Command{
	Use:   "deleteall",
	Short: "Delete all Processes for a Colony",
	Long:  "Delete all Processes for a Colony",
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

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		fmt.Print("WARNING!!! Are you sure you want to delete all process in the Colony This operation cannot be undone! (YES,no): ")

		reader := bufio.NewReader(os.Stdin)
		reply, _ := reader.ReadString('\n')
		if reply == "YES\n" {
			client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
			err = client.DeleteAllProcesses(ColonyID, ColonyPrvKey)
			CheckError(err)

			log.WithFields(log.Fields{"colonyID": ColonyID}).Info("Deleting all processes in Colony")
		} else {
			log.Info("Aborting ...")
		}
	},
}

var closeSuccessful = &cobra.Command{
	Use:   "close",
	Short: "Close a Process as Successful",
	Long:  "Close a Process as Successful",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		process, err := client.GetProcess(ProcessID, RuntimePrvKey)
		CheckError(err)

		err = client.CloseSuccessful(process.ID, RuntimePrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"processID": process.ID}).Info("Process closed as Successful")
	},
}

var closeFailed = &cobra.Command{
	Use:   "fail",
	Short: "Close a Process as Failed",
	Long:  "Close a Process as Failed",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		process, err := client.GetProcess(ProcessID, RuntimePrvKey)
		CheckError(err)

		err = client.CloseFailed(process.ID, RuntimePrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"processID": process.ID}).Info("Process closed as Failed")
	},
}
