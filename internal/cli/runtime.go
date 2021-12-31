package cli

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"colonies/pkg/security/crypto"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kataras/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	runtimeCmd.AddCommand(registerRuntimeCmd)
	runtimeCmd.AddCommand(lsRuntimesCmd)
	runtimeCmd.AddCommand(approveRuntimeCmd)
	runtimeCmd.AddCommand(rejectRuntimeCmd)
	rootCmd.AddCommand(runtimeCmd)

	runtimeCmd.PersistentFlags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	runtimeCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	runtimeCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Server HTTP port")

	registerRuntimeCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	registerRuntimeCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony Runtime")
	registerRuntimeCmd.MarkFlagRequired("spec")

	lsRuntimesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	approveRuntimeCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	approveRuntimeCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Colony Runtime Id")
	approveRuntimeCmd.MarkFlagRequired("runtimeid")

	rejectRuntimeCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	rejectRuntimeCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Colony Runtime Id")
	rejectRuntimeCmd.MarkFlagRequired("runtimeid")
}

var runtimeCmd = &cobra.Command{
	Use:   "runtime",
	Short: "Manage Colony Runtimes",
	Long:  "Manage Colony Runtimes",
}

var registerRuntimeCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new Runtime",
	Long:  "Register a new Runtime",
	Run: func(cmd *cobra.Command, args []string) {
		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		runtime, err := core.ConvertJSONToRuntime(string(jsonSpecBytes))
		CheckError(err)

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		crypto := crypto.CreateCrypto()

		prvKey, err := crypto.GeneratePrivateKey()
		CheckError(err)

		runtimeID, err := crypto.GenerateID(prvKey)
		CheckError(err)
		runtime.SetID(runtimeID)
		runtime.SetColonyID(ColonyID)

		err = keychain.AddPrvKey(runtimeID, prvKey)
		CheckError(err)

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		addedRuntime, err := client.AddRuntime(runtime, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println(addedRuntime.ID)
	},
}

var lsRuntimesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all Runtimes available in a Colony",
	Long:  "List all Runtimes available in a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		runtimesFromServer, err := client.GetRuntimes(ColonyID, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		if JSON {
			jsonString, err := core.ConvertRuntimeArrayToJSON(runtimesFromServer)
			CheckError(err)
			fmt.Println(jsonString)
			os.Exit(0)
		}

		var data [][]string
		for _, runtime := range runtimesFromServer {
			status := ""
			switch runtime.Status {
			case core.PENDING:
				status = "Pending"
			case core.APPROVED:
				status = "Approved"
			case core.REJECTED:
				status = "Rejected"
			default:
				status = "Unknown"
			}

			data = append(data, []string{runtime.ID, runtime.Name, status})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Status"})

		for _, v := range data {
			table.Append(v)
		}

		table.Render()
	},
}

var approveRuntimeCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve a Colony Runtime",
	Long:  "Approve a Colony Runtime",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		runtime, err := client.GetRuntime(RuntimeID, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		err = client.ApproveRuntime(runtime.ID, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println("Colony Runtime with Id <" + runtime.ID + "> is now approved")
	},
}

var rejectRuntimeCmd = &cobra.Command{
	Use:   "rject",
	Short: "Reject a Colony Runtime",
	Long:  "Reject a Colony Runtime",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		runtime, err := client.GetRuntime(RuntimeID, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		err = client.RejectRuntime(runtime.ID, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println("Colony Runtime with Id <" + runtime.ID + "> is now rejected")
	},
}
