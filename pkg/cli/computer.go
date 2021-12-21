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
	computerCmd.AddCommand(registerComputerCmd)
	computerCmd.AddCommand(lsComputersCmd)
	computerCmd.AddCommand(approveComputersCmd)
	computerCmd.AddCommand(disapproveComputersCmd)
	rootCmd.AddCommand(computerCmd)

	computerCmd.PersistentFlags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	computerCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	computerCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Server HTTP port")

	registerComputerCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	registerComputerCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony Computer")
	registerComputerCmd.MarkFlagRequired("spec")

	approveComputersCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	approveComputersCmd.Flags().StringVarP(&ComputerID, "computerid", "", "", "Colony Computer Id")
	approveComputersCmd.MarkFlagRequired("computerid")

	disapproveComputersCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	disapproveComputersCmd.Flags().StringVarP(&ComputerID, "computerid", "", "", "Colony Computer Id")
	disapproveComputersCmd.MarkFlagRequired("computerid")
}

var computerCmd = &cobra.Command{
	Use:   "computer",
	Short: "Manage Colony Computers",
	Long:  "Manage Colony Computers",
}

var registerComputerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new Computer",
	Long:  "Register a new Computer",
	Run: func(cmd *cobra.Command, args []string) {
		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		computer, err := core.ConvertJSONToComputer(string(jsonSpecBytes))
		CheckError(err)

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		prvKey, err := security.GeneratePrivateKey()
		CheckError(err)

		computerID, err := security.GenerateID(prvKey)
		CheckError(err)
		computer.SetID(computerID)
		computer.SetColonyID(ColonyID)

		err = keychain.AddPrvKey(computerID, prvKey)
		CheckError(err)

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		addedComputer, err := client.AddComputer(computer, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println(addedComputer.ID())
	},
}

var lsComputersCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all Computers available in a Colony",
	Long:  "List all Computers available in a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		computers, err := client.GetComputersByColonyID(ColonyID, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		jsonString, err := core.ConvertComputerArrayToJSON(computers)
		CheckError(err)

		fmt.Println(jsonString)
	},
}

var approveComputersCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve a Colony Computer",
	Long:  "Approve a Colony Computer",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		computer, err := client.GetComputerByID(ComputerID, ColonyID, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		err = client.ApproveComputer(computer, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println("Colony Computer with Id <" + computer.ID() + "> is now approved")
	},
}

var disapproveComputersCmd = &cobra.Command{
	Use:   "disapprove",
	Short: "Disapprove a Colony Computer",
	Long:  "Disapprove a Colony Computer",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		computer, err := client.GetComputerByID(ComputerID, ColonyID, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		err = client.RejectComputer(computer, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println("Colony Computer with Id <" + computer.ID() + "> is now disapproved")
	},
}
