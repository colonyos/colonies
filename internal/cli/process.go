package cli

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"fmt"
	"io/ioutil"
	"os"

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

	submitProcessCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Runtime Id")
	submitProcessCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colony or Runtime private key")
	submitProcessCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony process")
	approveRuntimesCmd.MarkFlagRequired("spec")

	listWaitingProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listWaitingProcessesCmd.MarkFlagRequired("colonyid")
	listWaitingProcessesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	listWaitingProcessesCmd.MarkFlagRequired("runtimeid")
	listWaitingProcessesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	listWaitingProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")

	listRunningProcessesCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Runtime Id")
	listRunningProcessesCmd.MarkFlagRequired("id")
	listRunningProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listRunningProcessesCmd.MarkFlagRequired("colonyid")
	listRunningProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colony or Runtime private key")
	listRunningProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")

	listSuccessfulProcessesCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Runtime Id")
	listSuccessfulProcessesCmd.MarkFlagRequired("id")
	listSuccessfulProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listSuccessfulProcessesCmd.MarkFlagRequired("colonyid")
	listSuccessfulProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colony or Runtime private key")
	listSuccessfulProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")

	listFailedProcessesCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Runtime Id")
	listFailedProcessesCmd.MarkFlagRequired("id")
	listFailedProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listFailedProcessesCmd.MarkFlagRequired("colonyid")
	listFailedProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colony or Runtime private key")
	listFailedProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")

	getProcessCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Runtime Id")
	getProcessCmd.MarkFlagRequired("id")
	getProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	getProcessCmd.MarkFlagRequired("colonyid")
	getProcessCmd.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	getProcessCmd.MarkFlagRequired("processid")
	getProcessCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colony or Runtime private key")

	assignProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	assignProcessCmd.MarkFlagRequired("colonyid")
	assignProcessCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	assignProcessCmd.MarkFlagRequired("runtimeid")
	assignProcessCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")

	markSuccessfull.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	markSuccessfull.MarkFlagRequired("colonyid")
	markSuccessfull.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	markSuccessfull.MarkFlagRequired("runtimeid")
	markSuccessfull.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	markSuccessfull.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	markSuccessfull.MarkFlagRequired("processid")

	markFailed.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	markFailed.MarkFlagRequired("colonyid")
	markFailed.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	markFailed.MarkFlagRequired("runtimeid")
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

		if PrvKey == "" {
			PrvKey, err = keychain.GetPrvKey(ID)
			CheckError(err)
		}

		addedProcess, err := client.PublishProcessSpec(processSpec, PrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println(addedProcess.ToJSON())
	},
}

var assignProcessCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a Process to a Runtime",
	Long:  "Assign a Process to a Runtime",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		process, err := client.AssignProcess(RuntimeID, ColonyID, RuntimePrvKey)
		if err != nil {
			fmt.Println("No process was assigned")
		} else {
			jsonString, err := process.ToJSON()
			CheckError(err)
			fmt.Println(jsonString)
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

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		processes, err := client.GetWaitingProcesses(RuntimeID, ColonyID, Count, RuntimePrvKey)
		CheckError(err)

		if len(processes) == 0 {
			fmt.Println("No Waiting Process found")
		} else {
			jsonString, err := core.ConvertProcessArrayToJSON(processes)
			CheckError(err)

			fmt.Println(jsonString)
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

		if PrvKey == "" {
			PrvKey, err = keychain.GetPrvKey(ID)
			CheckError(err)
		}

		processes, err := client.GetRunningProcesses(ColonyID, Count, PrvKey)
		CheckError(err)

		if len(processes) == 0 {
			fmt.Println("No Running Process found")
		} else {
			jsonString, err := core.ConvertProcessArrayToJSON(processes)
			CheckError(err)

			fmt.Println(jsonString)
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

		if PrvKey == "" {
			PrvKey, err = keychain.GetPrvKey(ID)
			CheckError(err)
		}

		processes, err := client.GetSuccessfulProcesses(ColonyID, Count, PrvKey)
		CheckError(err)

		if len(processes) == 0 {
			fmt.Println("No Successful Process found")
		} else {
			jsonString, err := core.ConvertProcessArrayToJSON(processes)
			CheckError(err)

			fmt.Println(jsonString)
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

		if PrvKey == "" {
			PrvKey, err = keychain.GetPrvKey(ID)
			CheckError(err)
		}

		processes, err := client.GetFailedProcesses(ColonyID, Count, PrvKey)
		CheckError(err)

		if len(processes) == 0 {
			fmt.Println("No Failed Process found")
		} else {
			jsonString, err := core.ConvertProcessArrayToJSON(processes)
			CheckError(err)

			fmt.Println(jsonString)
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

		if PrvKey == "" {
			PrvKey, err = keychain.GetPrvKey(ID)
			CheckError(err)
		}

		fmt.Println(ProcessID)
		fmt.Println(ColonyID)

		process, err := client.GetProcessByID(ProcessID, ColonyID, PrvKey)
		if process == nil { // TODO: better error handling
			fmt.Println("Process not found")
			os.Exit(-1)
		}

		fmt.Println(process.ToJSON())
	},
}

var markSuccessfull = &cobra.Command{
	Use:   "successful",
	Short: "Mark a Process as Successful",
	Long:  "Mark a Process as Successful",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		process, err := client.GetProcessByID(ProcessID, ColonyID, RuntimePrvKey)
		CheckError(err)

		err = client.MarkSuccessful(process, RuntimePrvKey)
		CheckError(err)

		fmt.Println("Process marked as Successful")
	},
}

var markFailed = &cobra.Command{
	Use:   "failed",
	Short: "Mark a Process as Failed",
	Long:  "Mark a Process as Failed",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		process, err := client.GetProcessByID(ProcessID, ColonyID, RuntimePrvKey)
		CheckError(err)

		err = client.MarkFailed(process, RuntimePrvKey)
		CheckError(err)

		fmt.Println("Process marked as Failed")
	},
}
