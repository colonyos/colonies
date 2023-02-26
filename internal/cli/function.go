package cli

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(functionCmd)
	functionCmd.AddCommand(execFuncCmd)
	functionCmd.AddCommand(addFuncCmd)
	functionCmd.AddCommand(removeFuncCmd)
	functionCmd.AddCommand(listFuncCmd)

	addFuncCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a function")
	addFuncCmd.MarkFlagRequired("spec")

	execFuncCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	execFuncCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	execFuncCmd.Flags().StringVarP(&TargetExecutorType, "targettype", "", "", "Target executor type")
	execFuncCmd.Flags().StringVarP(&TargetExecutorID, "targetid", "", "", "Target executor Id")
	execFuncCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	execFuncCmd.Flags().StringVarP(&Func, "func", "", "", "Remote function to call")
	execFuncCmd.Flags().StringSliceVarP(&Args, "args", "", make([]string, 0), "Arguments")
	execFuncCmd.Flags().StringSliceVarP(&Env, "env", "", make([]string, 0), "Environment")
	execFuncCmd.Flags().IntVarP(&MaxWaitTime, "maxwaittime", "", -1, "Maximum queue wait time")
	execFuncCmd.Flags().IntVarP(&MaxExecTime, "maxexectime", "", -1, "Maximum execution time in seconds before failing")
	execFuncCmd.Flags().IntVarP(&MaxRetries, "maxretries", "", -1, "Maximum number of retries when failing")
	execFuncCmd.Flags().BoolVarP(&Wait, "wait", "", false, "Colony Id")
}

var functionCmd = &cobra.Command{

	Use:   "function",
	Short: "Manage functions",
	Long:  "Manage functions",
}

var addFuncCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a function",
	Long:  "Add a function",
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

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		funcSpec, err := core.ConvertJSONToFunction(string(jsonSpecBytes))
		CheckError(err)

		funcSpec.ColonyID = ColonyID
		funcSpec.ExecutorID = ExecutorID

		addedFunc, err := client.AddFunction(funcSpec, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"FunctionID": addedFunc.FunctionID, "ExecutorID": ExecutorID, "ColonyID": ColonyID}).Info("Function added")
	},
}

var removeFuncCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a function",
	Long:  "Remove a function",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

	},
}

var listFuncCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all functions",
	Long:  "List all functions",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

	},
}

var execFuncCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a function",
	Long:  "Execute a function",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		env := make(map[string]string)
		for _, v := range Env {
			s := strings.Split(v, "=")
			if len(s) != 2 {
				CheckError(errors.New("Invalid key-value pair, try e.g. --env key1=value1,key2=value2 "))
			}
			key := s[0]
			value := s[1]
			env[key] = value
		}

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

		if TargetExecutorType == "" && TargetExecutorID == "" {
			CheckError(errors.New("Target Executor Type or Target Executor ID must be specified"))
		}

		var conditions core.Conditions
		if TargetExecutorType != "" {
			conditions = core.Conditions{ColonyID: ColonyID, ExecutorType: TargetExecutorType}
		} else {
			conditions = core.Conditions{ColonyID: ColonyID, ExecutorIDs: []string{TargetExecutorID}}
		}

		processSpec := core.ProcessSpec{
			Func:        Func,
			Args:        Args,
			MaxWaitTime: MaxWaitTime,
			MaxExecTime: MaxExecTime,
			MaxRetries:  MaxRetries,
			Conditions:  conditions,
			Env:         env}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		addedProcess, err := client.SubmitProcessSpec(&processSpec, ExecutorPrvKey)
		CheckError(err)

		if Wait {
			wait(client, addedProcess)
		} else {
			log.WithFields(log.Fields{"ProcessID": addedProcess.ID}).Info("Process submitted")
		}
	},
}
