package cli

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

func init() {
	processCmd.AddCommand(submitProcessCmd)
	processCmd.AddCommand(listWaitingProcessesCmd)
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
	listWaitingProcessesCmd.Flags().StringVarP(&ComputerPrvKey, "colonyprvkey", "", "", "Computer private key")
	listWaitingProcessesCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of processes to list")
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

		addedProcess, err := client.SubmitProcessSpec(processSpec, PrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println(addedProcess.ToJSON())
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

		jsonString, err := core.ConvertProcessArrayToJSON(processes)
		CheckError(err)

		fmt.Println(jsonString)
	},
}
