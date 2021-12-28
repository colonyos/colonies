package cli

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/kataras/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	processCmd.AddCommand(submitProcessCmd)
	processCmd.AddCommand(listWaitingProcessesCmd)
	processCmd.AddCommand(listRunningProcessesCmd)
	processCmd.AddCommand(listSuccessfulProcessesCmd)
	processCmd.AddCommand(listFailedProcessesCmd)
	processCmd.AddCommand(getProcessCmd)
	processCmd.AddCommand(assignProcessCmd)
	processCmd.AddCommand(markSuccessfull)
	processCmd.AddCommand(markFailed)
	rootCmd.AddCommand(processCmd)

	processCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	processCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Server HTTP port")

	submitProcessCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	submitProcessCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	submitProcessCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony process")
	submitProcessCmd.MarkFlagRequired("spec")

	listWaitingProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listWaitingProcessesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listWaitingProcessesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listWaitingProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")
	listWaitingProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	listRunningProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listRunningProcessesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listRunningProcessesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listRunningProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")
	listRunningProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	listSuccessfulProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listSuccessfulProcessesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listSuccessfulProcessesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listSuccessfulProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")
	listSuccessfulProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	listFailedProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listFailedProcessesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listFailedProcessesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listFailedProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")
	listFailedProcessesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	getProcessCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	getProcessCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	getProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	getProcessCmd.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	getProcessCmd.MarkFlagRequired("processid")
	getProcessCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	assignProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	assignProcessCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	assignProcessCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")

	markSuccessfull.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	markSuccessfull.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	markSuccessfull.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	markSuccessfull.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	markSuccessfull.MarkFlagRequired("processid")

	markFailed.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	markFailed.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	markFailed.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	markFailed.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	markFailed.MarkFlagRequired("processid")
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
		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		processSpec, err := core.ConvertJSONToProcessSpec(string(jsonSpecBytes))
		CheckError(err)

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		addedProcess, err := client.PublishProcessSpec(processSpec, RuntimePrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println(addedProcess.ID)
	},
}

var assignProcessCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a Process to a Runtime",
	Long:  "Assign a Process to a Runtime",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		process, err := client.AssignProcess(RuntimeID, ColonyID, RuntimePrvKey)
		if err != nil {
			fmt.Println("No process was assigned")
		} else {
			fmt.Println("Process with Id <" + process.ID + "> was assigned to Runtime with Id <" + RuntimeID + ">")
		}
	},
}

var listWaitingProcessesCmd = &cobra.Command{
	Use:   "psw",
	Short: "List all Waiting Processes",
	Long:  "List all Waiting Processes",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		processes, err := client.GetWaitingProcesses(RuntimeID, ColonyID, Count, RuntimePrvKey)
		CheckError(err)

		if len(processes) == 0 {
			fmt.Println("No Waiting Process found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.SubmissionTime.Format(TimeLayout)})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Submission time"})
			for _, v := range data {
				table.Append(v)
			}
			table.Render()
		}

	},
}

var listRunningProcessesCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all Running Processes",
	Long:  "List all Running Processes",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		processes, err := client.GetRunningProcesses(ColonyID, Count, RuntimePrvKey)
		CheckError(err)

		if len(processes) == 0 {
			fmt.Println("No nunning process found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.StartTime.Format(TimeLayout)})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Start time"})
			for _, v := range data {
				table.Append(v)
			}
			table.Render()
		}
	},
}

var listSuccessfulProcessesCmd = &cobra.Command{
	Use:   "pss",
	Short: "List all Successfull Processes",
	Long:  "List all Successfull Processes",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		processes, err := client.GetSuccessfulProcesses(ColonyID, Count, RuntimePrvKey)
		CheckError(err)

		if len(processes) == 0 {
			fmt.Println("No successful process found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.EndTime.Format(TimeLayout)})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "End time"})
			for _, v := range data {
				table.Append(v)
			}
			table.Render()
		}
	},
}

var listFailedProcessesCmd = &cobra.Command{
	Use:   "psf",
	Short: "List all Failed Processes",
	Long:  "List all Failed Processes",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		processes, err := client.GetFailedProcesses(ColonyID, Count, RuntimePrvKey)
		CheckError(err)

		if len(processes) == 0 {
			fmt.Println("No failed process found")
		} else {
			if JSON {
				jsonString, err := core.ConvertProcessArrayToJSON(processes)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			var data [][]string
			for _, process := range processes {
				data = append(data, []string{process.ID, process.EndTime.Format(TimeLayout)})
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "End time"})
			for _, v := range data {
				table.Append(v)
			}
			table.Render()
		}
	},
}

var getProcessCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a Process",
	Long:  "Get info about a Process",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		process, err := client.GetProcessByID(ProcessID, ColonyID, RuntimePrvKey)
		if process == nil {
			fmt.Println("Process with Id <" + process.ID + "> not found")
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

		var status string
		switch process.Status {
		case core.WAITING:
			status = "Waiting"
		case core.RUNNING:
			status = "Running"
		case core.SUCCESS:
			status = "Successful"
		case core.FAILED:
			status = "Failed"
		default:
			status = "Unkown"
		}

		fmt.Println("Process:")

		processData := [][]string{
			[]string{"ID", process.ID},
			[]string{"IsAssigned", isAssigned},
			[]string{"AssignedRuntimeID", assignedRuntimeID},
			[]string{"Status", status},
			[]string{"SubmissionTime", process.SubmissionTime.Format(TimeLayout)},
			[]string{"StartTime", process.StartTime.Format(TimeLayout)},
			[]string{"EndTime", process.EndTime.Format(TimeLayout)},
			[]string{"Deadline", process.Deadline.Format(TimeLayout)},
			[]string{"Retries", strconv.Itoa(process.Retries)},
		}
		processTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range processData {
			processTable.Append(v)
		}
		processTable.SetAlignment(tablewriter.ALIGN_LEFT)
		processTable.Render()

		fmt.Println()
		fmt.Println("Requirements:")

		runtimeIDs := ""
		for _, runtimeID := range process.ProcessSpec.Conditions.RuntimeIDs {
			runtimeIDs += runtimeID + "\n"
		}
		runtimeIDs = strings.TrimSuffix(runtimeIDs, "\n")

		if runtimeIDs == "" {
			runtimeIDs = "None"
		}

		specData := [][]string{
			[]string{"ColonyID", process.ProcessSpec.Conditions.ColonyID},
			[]string{"RuntimeIDs", runtimeIDs},
			[]string{"RuntimeType", process.ProcessSpec.Conditions.RuntimeType},
			[]string{"Memory", strconv.Itoa(process.ProcessSpec.Conditions.Mem)},
			[]string{"CPU Cores", strconv.Itoa(process.ProcessSpec.Conditions.Cores)},
			[]string{"Number of GPUs", strconv.Itoa(process.ProcessSpec.Conditions.GPUs)},
			[]string{"Timeout", strconv.Itoa(process.ProcessSpec.Timeout)},
			[]string{"Max retries", strconv.Itoa(process.ProcessSpec.MaxRetries)},
		}
		specTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range specData {
			specTable.Append(v)
		}
		specTable.SetAlignment(tablewriter.ALIGN_LEFT)
		specTable.Render()

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
			attributeData = append(attributeData, []string{attribute.ID, attribute.Key, attribute.Value, attributeType})
		}

		attributeTable := tablewriter.NewWriter(os.Stdout)
		attributeTable.SetHeader([]string{"ID", "Key", "Value", "Type"})
		for _, v := range attributeData {
			attributeTable.Append(v)
		}
		attributeTable.Render()
	},
}

var markSuccessfull = &cobra.Command{
	Use:   "successful",
	Short: "Mark a Process as Successful",
	Long:  "Mark a Process as Successful",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		process, err := client.GetProcessByID(ProcessID, ColonyID, RuntimePrvKey)
		CheckError(err)

		err = client.MarkSuccessful(process, RuntimePrvKey)
		CheckError(err)

		fmt.Println("Process with Id <" + process.ID + "> marked as successful")
	},
}

var markFailed = &cobra.Command{
	Use:   "failed",
	Short: "Mark a Process as Failed",
	Long:  "Mark a Process as Failed",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		process, err := client.GetProcessByID(ProcessID, ColonyID, RuntimePrvKey)
		CheckError(err)

		err = client.MarkFailed(process, RuntimePrvKey)
		CheckError(err)

		fmt.Println("Process with Id <" + process.ID + "> marked as failed")
	},
}
