package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	executorCmd.AddCommand(addExecutorCmd)
	executorCmd.AddCommand(removeExecutorCmd)
	executorCmd.AddCommand(lsExecutorsCmd)
	executorCmd.AddCommand(getExecutorCmd)
	executorCmd.AddCommand(approveExecutorCmd)
	executorCmd.AddCommand(rejectExecutorCmd)
	executorCmd.AddCommand(resolveExecutorCmd)
	executorCmd.AddCommand(osExecutorCmd)
	rootCmd.AddCommand(executorCmd)

	executorCmd.PersistentFlags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	executorCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	executorCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addExecutorCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	addExecutorCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	addExecutorCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	addExecutorCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of an executor")
	addExecutorCmd.Flags().StringVarP(&ExecutorName, "name", "", "", "Executor name")
	addExecutorCmd.Flags().StringVarP(&ExecutorType, "type", "", "", "Executor type")
	addExecutorCmd.Flags().BoolVarP(&Approve, "approve", "", false, "Also, approve the added executor")

	removeExecutorCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	removeExecutorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")

	lsExecutorsCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	lsExecutorsCmd.Flags().BoolVarP(&Full, "full", "", false, "Print detail info")
	lsExecutorsCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	lsExecutorsCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")

	getExecutorCmd.Flags().StringVarP(&TargetExecutorID, "targetid", "", "", "Target executor Id")
	getExecutorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	getExecutorCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")

	approveExecutorCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	approveExecutorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Colony Executor Id")
	approveExecutorCmd.MarkFlagRequired("executorid")

	rejectExecutorCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	rejectExecutorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	rejectExecutorCmd.MarkFlagRequired("executorid")

	resolveExecutorCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	resolveExecutorCmd.Flags().StringVarP(&TargetExecutorName, "executorname", "", "", "Executor name to resolve Id for")
	resolveExecutorCmd.MarkFlagRequired("executorid")
}

var executorCmd = &cobra.Command{
	Use:   "executor",
	Short: "Manage executors",
	Long:  "Manage executors",
}

var addExecutorCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new executor",
	Long:  "Add a new executor",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		crypto := crypto.CreateCrypto()
		var executorPrvKey string
		var err error
		if ExecutorPrvKey != "" {
			executorPrvKey = ExecutorPrvKey
			if len(executorPrvKey) != 64 {
				CheckError(errors.New("Invalid private key length"))
			}
		} else {
			executorPrvKey, err = crypto.GeneratePrivateKey()
			CheckError(err)
		}

		executorID, err := crypto.GenerateID(executorPrvKey)
		CheckError(err)

		var executor *core.Executor
		if SpecFile != "" {
			jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
			CheckError(err)
			executor, err = core.ConvertJSONToExecutor(string(jsonSpecBytes))
			CheckError(err)
		} else {
			if ExecutorName == "" {
				ExecutorName = os.Getenv("COLONIES_EXECUTOR_NAME")
			}

			if ExecutorName == "" {
				CheckError(errors.New("Executor name not specified"))
			}

			if os.Getenv("HOSTNAME") != "" {
				ExecutorName += "."
				ExecutorName += os.Getenv("HOSTNAME")
			}

			if ExecutorType == "" {
				ExecutorType = os.Getenv("COLONIES_EXECUTOR_TYPE")
			}

			if ExecutorType == "" {
				CheckError(errors.New("Executor type not specified"))
			}
			executor = core.CreateExecutor(executorID, ExecutorType, ExecutorName, ColonyID, time.Now(), time.Now())
		}

		executor.SetID(executorID)
		executor.SetColonyID(ColonyID)

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyPrvKey == "" {
			ColonyPrvKey = os.Getenv("COLONIES_COLONY_PRVKEY")
		}
		if ColonyPrvKey == "" {
			CheckError(errors.New("ERROR:" + ColonyPrvKey))
			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		addedExecutor, err := client.AddExecutor(executor, ColonyPrvKey)
		CheckError(err)

		err = keychain.AddPrvKey(executorID, executorPrvKey)
		CheckError(err)

		log.Info("Saving executor Id to /tmp/executorid")
		err = os.WriteFile("/tmp/executorid", []byte(executorID), 0644)
		CheckError(err)

		err = os.WriteFile("/tmp/executorprvkey", []byte(executorPrvKey), 0644)
		CheckError(err)
		log.Info("Saving executor prvKey to /tmp/executorprvkey")

		if Approve {
			log.WithFields(log.Fields{"ExecutorID": executorID}).Info("Approving Executor")
			err = client.ApproveExecutor(executorID, ColonyPrvKey)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ExecutorName": executor.Name, "ExecutorType": executor.Type, "ExecutorID": addedExecutor.ID, "ColonyID": ColonyID}).Info("Executor added")
	},
}

var removeExecutorCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an executor",
	Long:  "Remove an executor",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ColonyPrvKey == "" {
			ColonyPrvKey = os.Getenv("COLONIES_COLONY_PRVKEY")
		}
		if ColonyPrvKey == "" {
			keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
			CheckError(err)

			ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		if ExecutorID != "" {
			err := client.DeleteExecutor(ExecutorID, ColonyPrvKey)
			CheckError(err)
		} else {
			removeExecutorFromTmp(client)
		}

		log.WithFields(log.Fields{"ExecutorID": ExecutorID, "ColonyID": ColonyID}).Info("Executor removed")
	},
}

func printExecutor(client *client.ColoniesClient, executor *core.Executor) {
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

	requireFuncRegStr := "False"
	if executor.RequireFuncReg {
		requireFuncRegStr = "True"
	}

	fmt.Println("Executor:")

	executorData := [][]string{
		[]string{"Name", executor.Name},
		[]string{"ID", executor.ID},
		[]string{"Type", executor.Type},
		[]string{"ColonyID", executor.ColonyID},
		[]string{"State", state},
		[]string{"RequireFuncRegistration", requireFuncRegStr},
		[]string{"CommissionTime", executor.CommissionTime.Format(TimeLayout)},
		[]string{"LastHeardFrom", executor.LastHeardFromTime.Format(TimeLayout)},
	}

	executorTable := tablewriter.NewWriter(os.Stdout)
	for _, v := range executorData {
		executorTable.Append(v)
	}
	executorTable.SetAlignment(tablewriter.ALIGN_LEFT)
	executorTable.Render()

	functions, err := client.GetFunctionsByExecutorID(executor.ID, ExecutorPrvKey)
	CheckError(err)

	fmt.Println()
	fmt.Println("Functions:")

	if len(functions) == 0 {
		fmt.Println("No functions found")
	} else {
		for innerCounter, function := range functions {
			funcData := [][]string{
				[]string{"FuncName", function.FuncName},
				[]string{"FunctionId", function.FunctionID},
				[]string{"Counter", strconv.Itoa(function.Counter)},
				[]string{"MinWaitTime", fmt.Sprintf("%f s", function.MinWaitTime)},
				[]string{"MaxWaitTime", fmt.Sprintf("%f s", function.MaxWaitTime)},
				[]string{"AvgWaitTime", fmt.Sprintf("%f s", function.AvgWaitTime)},
				[]string{"MinExecTime", fmt.Sprintf("%f s", function.MinExecTime)},
				[]string{"MaxExecTime", fmt.Sprintf("%f s", function.MaxExecTime)},
				[]string{"AvgExecTime", fmt.Sprintf("%f s", function.AvgExecTime)},
			}
			funcTable := tablewriter.NewWriter(os.Stdout)
			for _, v := range funcData {
				funcTable.Append(v)
			}
			funcTable.SetAlignment(tablewriter.ALIGN_LEFT)
			funcTable.Render()
			if innerCounter != len(functions)-1 {
				fmt.Println()
			}
		}
	}
}

var lsExecutorsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all executors",
	Long:  "List all executors",
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
				printExecutor(client, executor)

				if counter != len(executorsFromServer)-1 {
					fmt.Println()
					fmt.Println("==============================================================================================")
					fmt.Println()
				} else {
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

var getExecutorCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about an executor",
	Long:  "Get info about an executor",
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

		executorFromServer, err := client.GetExecutor(TargetExecutorID, ExecutorPrvKey)
		if err != nil {
			// Try ColonyPrvKey instead
			if ColonyPrvKey == "" {
				if ColonyID == "" {
					ColonyID = os.Getenv("COLONIES_COLONY_ID")
				}
				ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
				CheckError(err)
			}
			executorFromServer, err = client.GetExecutor(TargetExecutorID, ColonyPrvKey)
			CheckError(err)
		}

		printExecutor(client, executorFromServer)
	},
}

var approveExecutorCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve an executor",
	Long:  "Approve an executor",
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

		log.WithFields(log.Fields{"ExecutorID": ExecutorID, "ColonyID": ColonyID}).Info("Executor approved")
	},
}

var rejectExecutorCmd = &cobra.Command{
	Use:   "reject",
	Short: "Reject an executor",
	Long:  "Reject an executor",
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

		log.WithFields(log.Fields{"ExecutorID": ExecutorID, "ColonyID": ColonyID}).Info("Executor rejected")
	},
}

func removeExecutorFromTmp(client *client.ColoniesClient) {
	mutex.Lock()
	defer mutex.Unlock()

	executorIDBytes, err := os.ReadFile("/tmp/executorid")
	CheckError(err)

	executorID := string(executorIDBytes)

	err = client.DeleteExecutor(executorID, ColonyPrvKey)
	CheckError(err)
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
