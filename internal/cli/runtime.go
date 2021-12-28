package cli

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
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
	runtimeCmd.AddCommand(approveRuntimesCmd)
	runtimeCmd.AddCommand(disapproveRuntimesCmd)
	rootCmd.AddCommand(runtimeCmd)

	runtimeCmd.PersistentFlags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	runtimeCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	runtimeCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Server HTTP port")

	registerRuntimeCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	registerRuntimeCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony Runtime")
	registerRuntimeCmd.MarkFlagRequired("spec")

	lsRuntimesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	approveRuntimesCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	approveRuntimesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Colony Runtime Id")
	approveRuntimesCmd.MarkFlagRequired("runtimeid")

	disapproveRuntimesCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	disapproveRuntimesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Colony Runtime Id")
	disapproveRuntimesCmd.MarkFlagRequired("runtimeid")
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

		prvKey, err := security.GeneratePrivateKey()
		CheckError(err)

		runtimeID, err := security.GenerateID(prvKey)
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

		runtimesFromServer, err := client.GetRuntimesByColonyID(ColonyID, ColonyPrvKey, ServerHost, ServerPort)
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
				status = "Disapproved"
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

var approveRuntimesCmd = &cobra.Command{
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

		runtime, err := client.GetRuntimeByID(RuntimeID, ColonyID, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		err = client.ApproveRuntime(runtime, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println("Colony Runtime with Id <" + runtime.ID + "> is now approved")
	},
}

var disapproveRuntimesCmd = &cobra.Command{
	Use:   "disapprove",
	Short: "Disapprove a Colony Runtime",
	Long:  "Disapprove a Colony Runtime",
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

		runtime, err := client.GetRuntimeByID(RuntimeID, ColonyID, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		err = client.RejectRuntime(runtime, ColonyPrvKey, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println("Colony Runtime with Id <" + runtime.ID + "> is now disapproved")
	},
}
