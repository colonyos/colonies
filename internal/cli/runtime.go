package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/kataras/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	runtimeCmd.AddCommand(registerRuntimeCmd)
	runtimeCmd.AddCommand(lsRuntimesCmd)
	runtimeCmd.AddCommand(approveRuntimeCmd)
	runtimeCmd.AddCommand(rejectRuntimeCmd)
	runtimeCmd.AddCommand(deleteRuntimeCmd)
	rootCmd.AddCommand(runtimeCmd)

	runtimeCmd.PersistentFlags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	runtimeCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	runtimeCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Server HTTP port")

	registerRuntimeCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	registerRuntimeCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony Runtime")
	registerRuntimeCmd.MarkFlagRequired("spec")

	lsRuntimesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	lsRuntimesCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	lsRuntimesCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")

	approveRuntimeCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	approveRuntimeCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Colony Runtime Id")
	approveRuntimeCmd.MarkFlagRequired("runtimeid")

	rejectRuntimeCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	rejectRuntimeCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Colony Runtime Id")
	rejectRuntimeCmd.MarkFlagRequired("runtimeid")

	deleteRuntimeCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	deleteRuntimeCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Colony Runtime Id")
	deleteRuntimeCmd.MarkFlagRequired("runtimeid")
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
		parseServerEnv()

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

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		addedRuntime, err := client.AddRuntime(runtime, ColonyPrvKey)
		CheckError(err)

		err = keychain.AddPrvKey(runtimeID, prvKey)
		CheckError(err)

		fmt.Println(addedRuntime.ID)
	},
}

var lsRuntimesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all Runtimes available in a Colony",
	Long:  "List all Runtimes available in a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimePrvKey == "" {
			if RuntimeID == "" {
				RuntimeID = os.Getenv("RUNTIMEID")
			}
			RuntimePrvKey, _ = keychain.GetPrvKey(RuntimeID)
		}

		if RuntimePrvKey == "" {
			if RuntimeID == "" {
				RuntimeID = os.Getenv("RUNTIMEID")
			}
			RuntimePrvKey, _ = keychain.GetPrvKey(RuntimeID)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		runtimesFromServer, err := client.GetRuntimes(ColonyID, RuntimePrvKey)
		prvKey := RuntimePrvKey
		if err != nil {
			// Try ColonyPrvKey instead
			if ColonyPrvKey == "" {
				if ColonyID == "" {
					ColonyID = os.Getenv("COLONYID")
				}
				ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
				CheckError(err)
			}
			runtimesFromServer, err = client.GetRuntimes(ColonyID, ColonyPrvKey)
			CheckError(err)
			prvKey = ColonyPrvKey
		}

		waitingTimes60 := make(map[string]float64)
		processingTimes60 := make(map[string]float64)
		utilizations60 := make(map[string]float64)
		for _, runtime := range runtimesFromServer {
			processes, err := client.GetProcessHistForRuntime(core.SUCCESS, ColonyID, runtime.ID, 60, prvKey)
			CheckError(err)

			waitingTimes60[runtime.ID] = utils.CalcAvgWaitingTime(processes)
			processingTimes60[runtime.ID] = utils.CalcAvgProcessingTime(processes)
			utilizations60[runtime.ID] = utils.CalcUtilization(processes)
		}

		waitingTimes3600 := make(map[string]float64)
		processingTimes3600 := make(map[string]float64)
		utilizations3600 := make(map[string]float64)
		for _, runtime := range runtimesFromServer {
			processes, err := client.GetProcessHistForRuntime(core.SUCCESS, ColonyID, runtime.ID, 3600, prvKey)
			CheckError(err)

			waitingTimes3600[runtime.ID] = utils.CalcAvgWaitingTime(processes)
			processingTimes3600[runtime.ID] = utils.CalcAvgProcessingTime(processes)
			utilizations3600[runtime.ID] = utils.CalcUtilization(processes)
		}

		if JSON {
			jsonString, err := core.ConvertRuntimeArrayToJSON(runtimesFromServer)
			CheckError(err)
			fmt.Println(jsonString)
			os.Exit(0)
		}

		for counter, runtime := range runtimesFromServer {
			state := ""
			switch runtime.State {
			case core.PENDING:
				state = "Pending"
			case core.APPROVED:
				state = "Approved"
			case core.REJECTED:
				state = "Rejected"
			default:
				state = "Unknown"
			}

			waitingTime60 := fmt.Sprintf("%f", waitingTimes60[runtime.ID])
			processingTime60 := fmt.Sprintf("%f", processingTimes60[runtime.ID])
			utilization60 := fmt.Sprintf("%f", utilizations60[runtime.ID])

			waitingTime3600 := fmt.Sprintf("%f", waitingTimes3600[runtime.ID])
			processingTime3600 := fmt.Sprintf("%f", processingTimes3600[runtime.ID])
			utilization3600 := fmt.Sprintf("%f", utilizations3600[runtime.ID])

			runtimeData := [][]string{
				[]string{"Name", runtime.Name},
				[]string{"ID", runtime.ID},
				[]string{"ColonyID", runtime.ColonyID},
				[]string{"State", state},
				[]string{"CommissionTime", runtime.CommissionTime.Format(TimeLayout)},
				[]string{"LastHeardFrom", runtime.LastHeardFromTime.Format(TimeLayout)},
				[]string{"AvgWaitingTime (minute)", waitingTime60 + "s"},
				[]string{"AvgProcessingTime (minute)", processingTime60 + "s"},
				[]string{"Utilization (minute)", utilization60 + "%"},
				[]string{"AvgWaitingTime (hour)", waitingTime3600 + "s"},
				[]string{"AvgProcessingTime (hour)", processingTime3600 + "s"},
				[]string{"Utilization (hour)", utilization3600 + "%"},
				[]string{"CPU", runtime.CPU},
				[]string{"Cores", strconv.Itoa(runtime.Cores)},
				[]string{"Mem [MiB]", strconv.Itoa(runtime.Mem)},
				[]string{"GPU", runtime.GPU},
				[]string{"GPUs", strconv.Itoa(runtime.GPUs)},
			}

			runtimeTable := tablewriter.NewWriter(os.Stdout)
			for _, v := range runtimeData {
				runtimeTable.Append(v)
			}
			runtimeTable.SetAlignment(tablewriter.ALIGN_LEFT)
			runtimeTable.Render()

			if counter != len(runtimesFromServer)-1 {
				fmt.Println()
			}
		}
	},
}

var approveRuntimeCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve a Colony Runtime",
	Long:  "Approve a Colony Runtime",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

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

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		err = client.ApproveRuntime(RuntimeID, ColonyPrvKey)
		CheckError(err)

		fmt.Println("Colony Runtime with Id <" + RuntimeID + "> is now approved")
	},
}

var rejectRuntimeCmd = &cobra.Command{
	Use:   "reject",
	Short: "Reject a Colony Runtime",
	Long:  "Reject a Colony Runtime",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

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

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		err = client.RejectRuntime(RuntimeID, ColonyPrvKey)
		CheckError(err)

		fmt.Println("Colony Runtime with Id <" + RuntimeID + "> is now rejected")
	},
}

var deleteRuntimeCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister a Colony Runtime",
	Long:  "Unregister a Colony Runtime",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

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

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		err = client.DeleteRuntime(RuntimeID, ColonyPrvKey)
		CheckError(err)

		fmt.Println("Colony Runtime with Id <" + RuntimeID + "> is now unregistered")
	},
}
