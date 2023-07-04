package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(functionCmd)
	functionCmd.AddCommand(execFuncCmd)
	functionCmd.AddCommand(submitFunctionSpecCmd)
	functionCmd.AddCommand(registerFuncCmd)
	functionCmd.AddCommand(removeFuncCmd)
	functionCmd.AddCommand(listFuncCmd)

	registerFuncCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a function")
	registerFuncCmd.MarkFlagRequired("spec")

	submitFunctionSpecCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	submitFunctionSpecCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	submitFunctionSpecCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a process")
	submitFunctionSpecCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	submitFunctionSpecCmd.Flags().BoolVarP(&Wait, "wait", "", false, "Wait for process to finish")
	submitFunctionSpecCmd.MarkFlagRequired("spec")
	submitFunctionSpecCmd.Flags().BoolVarP(&PrintOutput, "out", "", false, "Print process output, wait flag must be set")

	execFuncCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	execFuncCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	execFuncCmd.Flags().StringVarP(&TargetExecutorType, "targettype", "", "", "Target executor type")
	execFuncCmd.Flags().StringVarP(&TargetExecutorID, "targetid", "", "", "Target executor Id")
	execFuncCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	execFuncCmd.Flags().StringVarP(&FuncName, "func", "", "", "Remote function to call")
	execFuncCmd.Flags().StringSliceVarP(&Args, "args", "", make([]string, 0), "Arguments")
	execFuncCmd.Flags().StringSliceVarP(&Env, "env", "", make([]string, 0), "Environment")
	execFuncCmd.Flags().IntVarP(&MaxWaitTime, "maxwaittime", "", -1, "Maximum queue wait time")
	execFuncCmd.Flags().IntVarP(&MaxExecTime, "maxexectime", "", -1, "Maximum execution time in seconds before failing")
	execFuncCmd.Flags().IntVarP(&MaxRetries, "maxretries", "", -1, "Maximum number of retries when failing")
	execFuncCmd.Flags().BoolVarP(&Wait, "wait", "", false, "Wait for process to finish")
	execFuncCmd.Flags().BoolVarP(&PrintOutput, "out", "", false, "Print process output, wait flag must be set")

	removeFuncCmd.Flags().StringVarP(&FunctionID, "functionid", "", "", "FunctionID")
	removeFuncCmd.MarkFlagRequired("functionid")
}

var functionCmd = &cobra.Command{

	Use:   "function",
	Short: "Manage functions",
	Long:  "Manage functions",
}

var registerFuncCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a function to an executor",
	Long:  "Register a function to an executor",
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
	Short: "Remove a function from an executor, use 'colonies executor ls --full' to get the functionid",
	Long:  "Remove a function from an executor, use 'colonies executor ls --full' to get the functionid",
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

		err = client.DeleteFunction(FunctionID, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"FunctionId": ColonyID}).Info("Function removed")
	},
}

type statsEntry struct {
	callsCounter    int
	executorCounter int
	minWaitTime     float64
	maxWaitTime     float64
	minExecTime     float64
	maxExecTime     float64
	avgWaitTime     float64
	avgExecTime     float64
}

var listFuncCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all functions",
	Long:  "List all functions",
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

		functions, err := client.GetFunctionsByColonyID(ColonyID, ExecutorPrvKey)
		CheckError(err)

		statsMap := make(map[string]statsEntry)
		for _, function := range functions {
			e, ok := statsMap[function.FuncName]
			if ok {
				statsMap[function.FuncName] = statsEntry{
					callsCounter:    e.callsCounter + function.Counter,
					executorCounter: e.executorCounter + 1,
					minWaitTime:     math.Min(e.minWaitTime, function.MinWaitTime),
					maxWaitTime:     math.Max(e.maxWaitTime, function.MaxWaitTime),
					minExecTime:     math.Min(e.minExecTime, function.MinExecTime),
					maxExecTime:     math.Max(e.maxExecTime, function.MaxExecTime),
					avgWaitTime:     (e.avgWaitTime + function.AvgWaitTime) / 2.0,
					avgExecTime:     (e.avgExecTime + function.AvgExecTime) / 2.0,
				}
			} else {
				statsMap[function.FuncName] = statsEntry{
					callsCounter:    function.Counter,
					executorCounter: 1,
					minWaitTime:     function.MinWaitTime,
					maxWaitTime:     function.MaxWaitTime,
					minExecTime:     function.MinExecTime,
					maxExecTime:     function.MaxExecTime,
					avgWaitTime:     function.AvgWaitTime,
					avgExecTime:     function.AvgExecTime,
				}
			}
		}

		counter := 0
		for funcName, s := range statsMap {
			fmt.Println("Function:")

			funcData := [][]string{
				[]string{"FuncName", funcName},
				[]string{"Calls", strconv.Itoa(s.callsCounter)},
				[]string{"Served by", strconv.Itoa(s.executorCounter) + " executors"},
				[]string{"MinWaitTime", fmt.Sprintf("%f s", s.minWaitTime)},
				[]string{"MaxWaitTime", fmt.Sprintf("%f s", s.maxWaitTime)},
				[]string{"AvgWaitTime", fmt.Sprintf("%f s", s.avgWaitTime)},
				[]string{"MinExecTime", fmt.Sprintf("%f s", s.minExecTime)},
				[]string{"MaxExecTime", fmt.Sprintf("%f s", s.maxExecTime)},
				[]string{"AvgExecTime", fmt.Sprintf("%f s", s.avgExecTime)},
			}

			funcTable := tablewriter.NewWriter(os.Stdout)
			for _, v := range funcData {
				funcTable.Append(v)
			}
			funcTable.SetAlignment(tablewriter.ALIGN_LEFT)
			funcTable.Render()

			if counter != len(statsMap)-1 {
				fmt.Println()
			}
			counter++
		}

		if counter == 0 {
			log.WithFields(log.Fields{"ColonyId": ColonyID}).Info("No functions found")
		}

	},
}

var submitFunctionSpecCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a function specification",
	Long:  "Submit a function specification",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		funcSpec, err := core.ConvertJSONToFunctionSpec(string(jsonSpecBytes))
		CheckError(err)

		if funcSpec.Conditions.ColonyID == "" {
			if ColonyID == "" {
				ColonyID = os.Getenv("COLONIES_COLONY_ID")
			}
			if ColonyID == "" {
				CheckError(errors.New("Unknown Colony Id, please set COLONYID env variable or specify ColonyID in JSON file"))
			}

			funcSpec.Conditions.ColonyID = ColonyID
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

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		addedProcess, err := client.Submit(funcSpec, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessID": addedProcess.ID}).Info("Process submitted")
		if Wait {
			wait(client, addedProcess)
			process, err := client.GetProcess(addedProcess.ID, ExecutorPrvKey)
			CheckError(err)
			if process.State == core.FAILED {
				log.WithFields(log.Fields{"ProcessID": addedProcess.ID, "Error": process.Errors}).Error("Process failed")
				os.Exit(-1)
			} else if process.State == core.SUCCESS {
				log.WithFields(log.Fields{"ProcessID": addedProcess.ID}).Info("Process finished successfully")
			}
			if PrintOutput {
				fmt.Println(StrArr2Str(IfArr2StringArr(addedProcess.Output)))
			}
			os.Exit(0)
		}
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

		argsif := make([]interface{}, len(Args))
		for i, v := range Args {
			argsif[i] = v
		}

		funcSpec := core.FunctionSpec{
			FuncName:    FuncName,
			Args:        argsif,
			MaxWaitTime: MaxWaitTime,
			MaxExecTime: MaxExecTime,
			MaxRetries:  MaxRetries,
			Conditions:  conditions,
			Env:         env}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		addedProcess, err := client.Submit(&funcSpec, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessID": addedProcess.ID}).Info("Process submitted")
		if Wait {
			wait(client, addedProcess)
			process, err := client.GetProcess(addedProcess.ID, ExecutorPrvKey)
			CheckError(err)
			if process.State == core.FAILED {
				log.WithFields(log.Fields{"ProcessID": addedProcess.ID, "Error": process.Errors}).Error("Process failed")
				os.Exit(-1)
			} else if process.State == core.SUCCESS {
				log.WithFields(log.Fields{"ProcessID": addedProcess.ID}).Info("Process finished successfully")
			}
			if PrintOutput {
				fmt.Println(StrArr2Str(IfArr2StringArr(process.Output)))
			}
			os.Exit(0)
		}
	},
}
