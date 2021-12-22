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

	submitProcessCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Computer Id")
	submitProcessCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colony or Computer private key")
	submitProcessCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony process")
	approveComputersCmd.MarkFlagRequired("spec")

	listWaitingProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listWaitingProcessesCmd.MarkFlagRequired("colonyid")
	listWaitingProcessesCmd.Flags().StringVarP(&ComputerID, "computerid", "", "", "Computer Id")
	listWaitingProcessesCmd.MarkFlagRequired("computerid")
	listWaitingProcessesCmd.Flags().StringVarP(&ComputerPrvKey, "computerprvkey", "", "", "Computer private key")
	listWaitingProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")

	listRunningProcessesCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Computer Id")
	listRunningProcessesCmd.MarkFlagRequired("id")
	listRunningProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listRunningProcessesCmd.MarkFlagRequired("colonyid")
	listRunningProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colony or Computer private key")
	listRunningProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")

	listSuccessfulProcessesCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Computer Id")
	listSuccessfulProcessesCmd.MarkFlagRequired("id")
	listSuccessfulProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listSuccessfulProcessesCmd.MarkFlagRequired("colonyid")
	listSuccessfulProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colony or Computer private key")
	listSuccessfulProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")

	listFailedProcessesCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Computer Id")
	listFailedProcessesCmd.MarkFlagRequired("id")
	listFailedProcessesCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	listFailedProcessesCmd.MarkFlagRequired("colonyid")
	listFailedProcessesCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colony or Computer private key")
	listFailedProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")

	getProcessCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Computer Id")
	getProcessCmd.MarkFlagRequired("id")
	getProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	getProcessCmd.MarkFlagRequired("colonyid")
	getProcessCmd.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	getProcessCmd.MarkFlagRequired("processid")
	getProcessCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Colony or Computer private key")

	assignProcessCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	assignProcessCmd.MarkFlagRequired("colonyid")
	assignProcessCmd.Flags().StringVarP(&ComputerID, "computerid", "", "", "Computer Id")
	assignProcessCmd.MarkFlagRequired("computerid")
	assignProcessCmd.Flags().StringVarP(&ComputerPrvKey, "computerprvkey", "", "", "Computer private key")

	markSuccessfull.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	markSuccessfull.MarkFlagRequired("colonyid")
	markSuccessfull.Flags().StringVarP(&ComputerID, "computerid", "", "", "Computer Id")
	markSuccessfull.MarkFlagRequired("computerid")
	markSuccessfull.Flags().StringVarP(&ComputerPrvKey, "computerprvkey", "", "", "Computer private key")
	markSuccessfull.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	markSuccessfull.MarkFlagRequired("processid")

	markFailed.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	markFailed.MarkFlagRequired("colonyid")
	markFailed.Flags().StringVarP(&ComputerID, "computerid", "", "", "Computer Id")
	markFailed.MarkFlagRequired("computerid")
	markFailed.Flags().StringVarP(&ComputerPrvKey, "computerprvkey", "", "", "Computer private key")
	markFailed.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	markFailed.MarkFlagRequired("processid")
}

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Manage Colony processes",
	Long:  "Manage Colony processes",
}

var submitProcessCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a process to a Colony",
	Long:  "Submit a process to a Colony",
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
	Short: "Assign a process to a Computer",
	Long:  "Assign a process to a Computer",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ComputerPrvKey == "" {
			ComputerPrvKey, err = keychain.GetPrvKey(ComputerID)
			CheckError(err)
		}

		process, err := client.AssignProcess(ComputerID, ColonyID, ComputerPrvKey)
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
	Short: "List all waiting processes",
	Long:  "List all waiting processes",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ComputerPrvKey == "" {
			ComputerPrvKey, err = keychain.GetPrvKey(ComputerID)
			CheckError(err)
		}

		processes, err := client.GetWaitingProcesses(ComputerID, ColonyID, Count, ComputerPrvKey)
		CheckError(err)

		if len(processes) == 0 {
			fmt.Println("No waiting processes found")
		} else {
			jsonString, err := core.ConvertProcessArrayToJSON(processes)
			CheckError(err)

			fmt.Println(jsonString)
		}

	},
}

var listRunningProcessesCmd = &cobra.Command{
	Use:   "ps",
	Short: "List all running processes",
	Long:  "List all running processes",
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
			fmt.Println("No running processes found")
		} else {
			jsonString, err := core.ConvertProcessArrayToJSON(processes)
			CheckError(err)

			fmt.Println(jsonString)
		}
	},
}

var listSuccessfulProcessesCmd = &cobra.Command{
	Use:   "pss",
	Short: "List all successfull processes",
	Long:  "List all successful processes",
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
			fmt.Println("No successful processes found")
		} else {
			jsonString, err := core.ConvertProcessArrayToJSON(processes)
			CheckError(err)

			fmt.Println(jsonString)
		}
	},
}

var listFailedProcessesCmd = &cobra.Command{
	Use:   "psf",
	Short: "List all failed processes",
	Long:  "List all failed processes",
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
			fmt.Println("No failed processes found")
		} else {
			jsonString, err := core.ConvertProcessArrayToJSON(processes)
			CheckError(err)

			fmt.Println(jsonString)
		}
	},
}

var getProcessCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a process",
	Long:  "Get info about a process",
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
	Short: "Mark a process as successful",
	Long:  "Mark a process as successful",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ComputerPrvKey == "" {
			ComputerPrvKey, err = keychain.GetPrvKey(ComputerID)
			CheckError(err)
		}

		process, err := client.GetProcessByID(ProcessID, ColonyID, ComputerPrvKey)
		CheckError(err)

		err = client.MarkSuccessful(process, ComputerPrvKey)
		CheckError(err)

		fmt.Println("Process marked as successful")
	},
}

var markFailed = &cobra.Command{
	Use:   "failed",
	Short: "Mark a process as failed",
	Long:  "Mark a process as failed",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ComputerPrvKey == "" {
			ComputerPrvKey, err = keychain.GetPrvKey(ComputerID)
			CheckError(err)
		}

		process, err := client.GetProcessByID(ProcessID, ColonyID, ComputerPrvKey)
		CheckError(err)

		err = client.MarkFailed(process, ComputerPrvKey)
		CheckError(err)

		fmt.Println("Process marked as failed")
	},
}
