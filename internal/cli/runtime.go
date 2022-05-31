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
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
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
	runtimeCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 50080, "Server HTTP port")

	registerRuntimeCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	registerRuntimeCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	registerRuntimeCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony Runtime")
	registerRuntimeCmd.MarkFlagRequired("spec")

	lsRuntimesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	lsRuntimesCmd.Flags().BoolVarP(&Full, "full", "", false, "Print detail info")
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
	Short: "Manage runtimes",
	Long:  "Manage runtimes",
}

var registerRuntimeCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new runtime",
	Long:  "Register a new runtime",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		runtime, err := core.ConvertJSONToRuntime(string(jsonSpecBytes))
		CheckError(err)

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		crypto := crypto.CreateCrypto()

		var prvKey string
		if RuntimePrvKey != "" {
			prvKey = RuntimePrvKey
			if len(prvKey) != 64 {
				CheckError(errors.New("invalid private key length"))
			}
		} else {
			prvKey, err = crypto.GeneratePrivateKey()
			CheckError(err)
		}

		runtimeID, err := crypto.GenerateID(prvKey)
		CheckError(err)
		runtime.SetID(runtimeID)
		runtime.SetColonyID(ColonyID)

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		addedRuntime, err := client.AddRuntime(runtime, ColonyPrvKey)
		CheckError(err)

		err = keychain.AddPrvKey(runtimeID, prvKey)
		CheckError(err)

		log.WithFields(log.Fields{"runtimeID": addedRuntime.ID, "colonyID": ColonyID}).Info("Runtime registered")
	},
}

var lsRuntimesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all runtimes available in a colony",
	Long:  "List all runtimes available in a colony",
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

		if RuntimePrvKey == "" {
			if RuntimeID == "" {
				RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
			}
			RuntimePrvKey, _ = keychain.GetPrvKey(RuntimeID)
		}

		if RuntimePrvKey == "" {
			if RuntimeID == "" {
				RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
			}
			RuntimePrvKey, _ = keychain.GetPrvKey(RuntimeID)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		runtimesFromServer, err := client.GetRuntimes(ColonyID, RuntimePrvKey)
		if err != nil {
			// Try ColonyPrvKey instead
			if ColonyPrvKey == "" {
				if ColonyID == "" {
					ColonyID = os.Getenv("COLONIES_COLONYID")
				}
				ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
				CheckError(err)
			}
			runtimesFromServer, err = client.GetRuntimes(ColonyID, ColonyPrvKey)
			CheckError(err)
		}

		if Full {
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

				runtimeData := [][]string{
					[]string{"Name", runtime.Name},
					[]string{"ID", runtime.ID},
					[]string{"Type", runtime.RuntimeType},
					[]string{"Group", runtime.RuntimeGroup},
					[]string{"ColonyID", runtime.ColonyID},
					[]string{"State", state},
					[]string{"CommissionTime", runtime.CommissionTime.Format(TimeLayout)},
					[]string{"LastHeardFrom", runtime.LastHeardFromTime.Format(TimeLayout)},
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
		} else {
			var data [][]string
			for _, runtime := range runtimesFromServer {
				data = append(data, []string{runtime.ID, runtime.Name, runtime.RuntimeType})
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Name", "Type"})

			for _, v := range data {
				table.Append(v)
			}

			table.Render()

		}
	},
}

var approveRuntimeCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve a colony runtime",
	Long:  "Approve a colony runtime",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		err = client.ApproveRuntime(RuntimeID, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"runtimeID": RuntimeID, "colonyID": ColonyID}).Info("Runtime approved")
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
			ColonyID = os.Getenv("COLONIES_COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		err = client.RejectRuntime(RuntimeID, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"runtimeID": RuntimeID, "colonyID": ColonyID}).Info("Runtime rejected")
	},
}

var deleteRuntimeCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister a colony runtime",
	Long:  "Unregister a colony runtime",
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		err = client.DeleteRuntime(RuntimeID, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"runtimeID": RuntimeID, "colonyID": ColonyID}).Info("Runtime unregistered")
	},
}
