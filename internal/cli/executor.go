package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	executorCmd.AddCommand(registerExecutorCmd)
	executorCmd.AddCommand(lsExecutorsCmd)
	executorCmd.AddCommand(approveExecutorCmd)
	executorCmd.AddCommand(rejectExecutorCmd)
	executorCmd.AddCommand(deleteExecutorCmd)
	executorCmd.AddCommand(resolveExecutorCmd)
	executorCmd.AddCommand(workerStartCmd)
	executorCmd.AddCommand(workerRegisterCmd)
	executorCmd.AddCommand(workerUnregisterCmd)
	rootCmd.AddCommand(executorCmd)

	executorCmd.PersistentFlags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	executorCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	executorCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	registerExecutorCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	registerExecutorCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	registerExecutorCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony Executor")
	registerExecutorCmd.MarkFlagRequired("spec")

	lsExecutorsCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	lsExecutorsCmd.Flags().BoolVarP(&Full, "full", "", false, "Print detail info")
	lsExecutorsCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	lsExecutorsCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")

	approveExecutorCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	approveExecutorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Colony Executor Id")
	approveExecutorCmd.MarkFlagRequired("executorid")

	rejectExecutorCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	rejectExecutorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	rejectExecutorCmd.MarkFlagRequired("executorid")

	deleteExecutorCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	deleteExecutorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	deleteExecutorCmd.MarkFlagRequired("executorid")

	resolveExecutorCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	resolveExecutorCmd.Flags().StringVarP(&TargetExecutorName, "targetname", "", "", "Target executor Id")
	resolveExecutorCmd.MarkFlagRequired("executorid")
}

var executorCmd = &cobra.Command{
	Use:   "executor",
	Short: "Manage executors",
	Long:  "Manage executors",
}

var registerExecutorCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new executor",
	Long:  "Register a new executor",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		executor, err := core.ConvertJSONToExecutor(string(jsonSpecBytes))
		CheckError(err)

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		crypto := crypto.CreateCrypto()

		var prvKey string
		if ExecutorPrvKey != "" {
			prvKey = ExecutorPrvKey
			if len(prvKey) != 64 {
				CheckError(errors.New("Invalid private key length"))
			}
		} else {
			prvKey, err = crypto.GeneratePrivateKey()
			CheckError(err)
		}

		executorID, err := crypto.GenerateID(prvKey)
		CheckError(err)
		executor.SetID(executorID)
		executor.SetColonyID(ColonyID)

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		addedExecutor, err := client.AddExecutor(executor, ColonyPrvKey)
		CheckError(err)

		err = keychain.AddPrvKey(executorID, prvKey)
		CheckError(err)

		log.WithFields(log.Fields{"executorID": addedExecutor.ID, "colonyID": ColonyID}).Info("Executor registered")
	},
}

var lsExecutorsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all executors available in a colony",
	Long:  "List all executors available in a colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorPrvKey == "" {
			if ExecutorID == "" {
				ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
			}
			ExecutorPrvKey, _ = keychain.GetPrvKey(ExecutorID)
		}

		if ExecutorPrvKey == "" {
			if ExecutorID == "" {
				ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
			}
			ExecutorPrvKey, _ = keychain.GetPrvKey(ExecutorID)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		executorsFromServer, err := client.GetExecutors(ColonyID, ExecutorPrvKey)
		if err != nil {
			// Try ColonyPrvKey instead
			if ColonyPrvKey == "" {
				if ColonyID == "" {
					ColonyID = os.Getenv("COLONIES_COLONY_ID")
				}
				ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
				CheckError(err)
			}
			executorsFromServer, err = client.GetExecutors(ColonyID, ColonyPrvKey)
			CheckError(err)
		}

		if Full {
			if JSON {
				jsonString, err := core.ConvertExecutorArrayToJSON(executorsFromServer)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			for counter, executor := range executorsFromServer {
				state := ""
				switch executor.State {
				case core.PENDING:
					state = "Pending"
				case core.APPROVED:
					state = "Approved"
				case core.REJECTED:
					state = "Rejected"
				default:
					state = "Unknown"
				}

				executorData := [][]string{
					[]string{"Name", executor.Name},
					[]string{"ID", executor.ID},
					[]string{"Type", executor.Type},
					[]string{"ColonyID", executor.ColonyID},
					[]string{"State", state},
					[]string{"CommissionTime", executor.CommissionTime.Format(TimeLayout)},
					[]string{"LastHeardFrom", executor.LastHeardFromTime.Format(TimeLayout)},
				}

				executorTable := tablewriter.NewWriter(os.Stdout)
				for _, v := range executorData {
					executorTable.Append(v)
				}
				executorTable.SetAlignment(tablewriter.ALIGN_LEFT)
				executorTable.Render()

				if counter != len(executorsFromServer)-1 {
					fmt.Println()
				}
			}
		} else {
			var data [][]string
			for _, executor := range executorsFromServer {
				data = append(data, []string{executor.ID, executor.Name, executor.Type})
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

var approveExecutorCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve a colony executor",
	Long:  "Approve a colony executor",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		err = client.ApproveExecutor(ExecutorID, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"executorID": ExecutorID, "colonyID": ColonyID}).Info("Executor approved")
	},
}

var rejectExecutorCmd = &cobra.Command{
	Use:   "reject",
	Short: "Reject a executor",
	Long:  "Reject a executor",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		err = client.RejectExecutor(ExecutorID, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"executorID": ExecutorID, "colonyID": ColonyID}).Info("Executor rejected")
	},
}

var deleteExecutorCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister a colony executor",
	Long:  "Unregister a colony executor",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}
		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		err = client.DeleteExecutor(ExecutorID, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"executorID": ExecutorID, "colonyID": ColonyID}).Info("Executor unregistered")
	},
}

var resolveExecutorCmd = &cobra.Command{
	Use:   "resolve",
	Short: "Resolve executor Id",
	Long:  "Resolve executor Id",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id, please set COLONYID env variable or specify ColonyID in JSON file"))
		}

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ExecutorID == "" {
			ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
		}
		if ExecutorID == "" {
			CheckError(errors.New("Unknown Executor Id"))
		}

		if ExecutorPrvKey == "" {
			ExecutorPrvKey, err = keychain.GetPrvKey(ExecutorID)
			CheckError(err)
		}

		if TargetExecutorName == "" {
			CheckError(errors.New("Target Executor Name must be specified"))
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		executors, err := client.GetExecutors(ColonyID, ExecutorPrvKey)
		CheckError(err)

		for _, executor := range executors {
			if executor.Name == TargetExecutorName {
				fmt.Println(executor.ID)
				os.Exit(0)
			}
		}

		log.WithFields(log.Fields{"ColonyId": ColonyID, "TargetExecutorName": TargetExecutorName}).Error("No such executor found")
	},
}
