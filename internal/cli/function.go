package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
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

	registerFuncCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	registerFuncCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")
	registerFuncCmd.MarkFlagRequired("name")
	registerFuncCmd.Flags().StringVarP(&FuncName, "func", "", "", "Function name to register")
	registerFuncCmd.MarkFlagRequired("func")

	submitFunctionSpecCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	submitFunctionSpecCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a process")
	submitFunctionSpecCmd.Flags().BoolVarP(&Wait, "wait", "", false, "Wait for process to finish")
	submitFunctionSpecCmd.MarkFlagRequired("spec")
	submitFunctionSpecCmd.Flags().BoolVarP(&PrintOutput, "out", "", false, "Print process output, wait flag must be set")
	submitFunctionSpecCmd.Flags().BoolVarP(&Follow, "follow", "", false, "Follow process, wait flag cannot be set")

	execFuncCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	execFuncCmd.Flags().StringVarP(&TargetExecutorType, "targettype", "", "", "Target executor type")
	execFuncCmd.Flags().StringVarP(&TargetExecutorID, "targetid", "", "", "Target executor Id")
	execFuncCmd.Flags().StringVarP(&FuncName, "func", "", "", "Remote function to call")
	execFuncCmd.Flags().StringSliceVarP(&Args, "args", "", make([]string, 0), "Arguments")
	execFuncCmd.Flags().StringSliceVarP(&Env, "env", "", make([]string, 0), "Environment")
	execFuncCmd.Flags().StringSliceVarP(&KwArgs, "kwargs", "", make([]string, 0), "Environment")
	execFuncCmd.Flags().StringSliceVarP(&Snapshots, "snapshots", "", make([]string, 0), "Environment")
	execFuncCmd.Flags().IntVarP(&MaxWaitTime, "maxwaittime", "", -1, "Maximum queue wait time")
	execFuncCmd.Flags().IntVarP(&MaxExecTime, "maxexectime", "", -1, "Maximum execution time in seconds before failing")
	execFuncCmd.Flags().IntVarP(&MaxRetries, "maxretries", "", -1, "Maximum number of retries when failing")
	execFuncCmd.Flags().BoolVarP(&Wait, "wait", "", false, "Wait for process to finish")
	execFuncCmd.Flags().BoolVarP(&PrintOutput, "out", "", false, "Print process output, wait flag must be set")
	execFuncCmd.Flags().BoolVarP(&Follow, "follow", "", false, "Follow process, wait flag cannot be set")

	removeFuncCmd.Flags().StringVarP(&FunctionID, "functionid", "", "", "FunctionID")
	removeFuncCmd.MarkFlagRequired("functionid")
}

var functionCmd = &cobra.Command{
	Use:   "function",
	Short: "Manage Functions",
	Long:  "Manage Functions",
}

var registerFuncCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a Function to an Executor",
	Long:  "Register a Function to an Executor",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if TargetExecutorName == "" {
			CheckError(errors.New("Executor name must be specified"))
		}

		if FuncName == "" {
			CheckError(errors.New("Func must be specified"))
		}

		f := &core.Function{ExecutorName: TargetExecutorName, ColonyName: ColonyName, FuncName: FuncName}
		regF, err := client.AddFunction(f, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"FunctionID": regF.FunctionID, "ExecutorName": regF.ExecutorName, "ColonyName": ColonyName, "FuncName": FuncName}).Info("Function added")
	},
}

var removeFuncCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a Function from an Executor  Hint: use 'colonies executor ls --full' to get the functionid",
	Long:  "Remove a Function from an Executor  Hint: use 'colonies executor ls --full' to get the functionid",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		err := client.RemoveFunction(FunctionID, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ColonyName": ColonyName, "FunctionId": FunctionID}).Info("Function removed")
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
	Short: "List all Functions",
	Long:  "List all Functions",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		functions, err := client.GetFunctionsByColonyName(ColonyName, PrvKey)
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
			log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("No functions found")
		}

	},
}

func follow(client *client.ColoniesClient, process *core.Process) {
	log.WithFields(log.Fields{"ProcessId": process.ID}).Info("Printing logs from process")
	var lastTimestamp int64
	lastTimestamp = 0
	for {
		logs, err := client.GetLogsByProcessIDSince(process.ID, Count, lastTimestamp, PrvKey)
		CheckError(err)

		process, err := client.GetProcess(process.ID, PrvKey)
		CheckError(err)

		if len(logs) == 0 {
			time.Sleep(500 * time.Millisecond)
			if process.State == core.SUCCESS {
				log.WithFields(log.Fields{"ProcessId": process.ID}).Info("Process finished successfully")
				os.Exit(0)
			}
			if process.State == core.FAILED {
				fmt.Println()
				log.WithFields(log.Fields{"ProcessId": process.ID}).Error("Process failed")
				os.Exit(-1)
			}
			continue
		} else {
			for _, log := range logs {
				fmt.Print(log.Message)
			}
			lastTimestamp = logs[len(logs)-1].Timestamp
		}

	}
}

func createSnapshot(funcSpec *core.FunctionSpec, client *client.ColoniesClient) {
	if len(funcSpec.Filesystem.SnapshotMounts) > 0 {
		for i, snapshotMount := range funcSpec.Filesystem.SnapshotMounts {
			snapshotName := core.GenerateRandomID()
			snapshot, err := client.CreateSnapshot(ColonyName, snapshotMount.Label, snapshotName, PrvKey)
			CheckError(err)
			funcSpec.Filesystem.SnapshotMounts[i].SnapshotID = snapshot.ID
			log.WithFields(log.Fields{"SnapshotID": snapshot.ID, "Label": snapshotMount.Label}).Debug("Creating snapshot")
		}
	}

}

var submitFunctionSpecCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a Function specification",
	Long:  "Submit a Function specification",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		funcSpec, err := core.ConvertJSONToFunctionSpec(string(jsonSpecBytes))
		if err != nil {
			if strings.Contains(err.Error(), "cannot unmarshal array into Go value of type core.FunctionSpec") {
				jsonStr := "{\"functionspecs\":" + string(jsonSpecBytes) + "}"
				_, err := core.ConvertJSONToWorkflowSpec(jsonStr)
				if err == nil {
					CheckError(errors.New("It looks like you are trying to submit a workflow, try to use colonies workflow submit --spec instead"))
				}
			}
		}
		CheckJSONParseErr(err, string(jsonSpecBytes))

		if funcSpec.Conditions.ColonyName == "" {
			funcSpec.Conditions.ColonyName = ColonyName
		}

		createSnapshot(funcSpec, client)

		addedProcess, err := client.Submit(funcSpec, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessId": addedProcess.ID}).Info("Process submitted")
		if Wait {
			wait(client, addedProcess)
			process, err := client.GetProcess(addedProcess.ID, PrvKey)
			CheckError(err)
			if process.State == core.FAILED {
				log.WithFields(log.Fields{"ProcessId": addedProcess.ID, "Error": process.Errors}).Error("Process failed")
				os.Exit(-1)
			} else if process.State == core.SUCCESS {
				log.WithFields(log.Fields{"ProcessId": addedProcess.ID}).Info("Process finished successfully")
			}
			if PrintOutput {
				fmt.Println(StrArr2Str(IfArr2StringArr(addedProcess.Output)))
			}
			os.Exit(0)
		} else if Follow {
			follow(client, addedProcess)
		}
	},
}

var execFuncCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a Function",
	Long:  "Execute a Function",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		env := make(map[string]string)
		for _, v := range Env {
			s := strings.Split(v, "=")
			if len(s) != 2 {
				CheckError(errors.New("Invalid key-value pair, try e.g. --env key1=value1,key2=value2"))
			}
			key := s[0]
			value := s[1]
			env[key] = value
		}

		kwargsIf := make(map[string]interface{})
		for _, v := range KwArgs {
			s := strings.Split(v, ":")
			if len(s) != 2 {
				CheckError(errors.New("Invalid key-value pair, try e.g. --kwargs cmd:python3,args:/tmp/xor/xor.py"))
			}
			key := s[0]
			value := s[1]

			if key == "args" {
				args := strings.Split(value, ",")
				argsif := make([]interface{}, len(args))
				for i, v := range args {
					argsif[i] = v
				}
				kwargsIf[key] = argsif
			} else {
				kwargsIf[key] = value
			}
		}

		if TargetExecutorType == "" && TargetExecutorID == "" {
			CheckError(errors.New("Target Executor Type or Target Executor ID must be specified"))
		}

		var conditions core.Conditions
		if TargetExecutorType != "" {
			conditions = core.Conditions{ColonyName: ColonyName, ExecutorType: TargetExecutorType}
		} else {
			conditions = core.Conditions{ColonyName: ColonyName, ExecutorNames: []string{TargetExecutorName}}
		}

		argsif := make([]interface{}, len(Args))
		for i, v := range Args {
			argsif[i] = v
		}

		funcSpec := core.FunctionSpec{
			FuncName:    FuncName,
			Args:        argsif,
			KwArgs:      kwargsIf,
			MaxWaitTime: MaxWaitTime,
			MaxExecTime: MaxExecTime,
			MaxRetries:  MaxRetries,
			Conditions:  conditions,
			Env:         env}

		createSnapshot(&funcSpec, client)

		addedProcess, err := client.Submit(&funcSpec, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessId": addedProcess.ID}).Info("Process submitted")
		if Wait {
			wait(client, addedProcess)
			process, err := client.GetProcess(addedProcess.ID, PrvKey)
			CheckError(err)
			if process.State == core.FAILED {
				log.WithFields(log.Fields{"ProcessId": addedProcess.ID, "Error": process.Errors}).Error("Process failed")
				os.Exit(-1)
			} else if process.State == core.SUCCESS {
				log.WithFields(log.Fields{"ProcessId": addedProcess.ID}).Info("Process finished successfully")
			}
			if PrintOutput {
				fmt.Println(StrArr2Str(IfArr2StringArr(process.Output)))
			}
			os.Exit(0)
		} else if Follow {
			follow(client, addedProcess)
		}
	},
}
