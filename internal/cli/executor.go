package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
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
	rootCmd.AddCommand(executorCmd)

	executorCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	executorCmd.PersistentFlags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	executorCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	executorCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addExecutorCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of an executor")
	addExecutorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor ID")
	addExecutorCmd.MarkFlagRequired("executorid")
	addExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")
	addExecutorCmd.Flags().StringVarP(&TargetExecutorType, "type", "", "", "Executor type")
	addExecutorCmd.Flags().BoolVarP(&Approve, "approve", "", false, "Also, approve the Executor")

	removeExecutorCmd.Flags().StringVarP(&TargetExecutorID, "executorid", "", "", "Executor Id")

	lsExecutorsCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	lsExecutorsCmd.Flags().BoolVarP(&Full, "full", "", false, "Print detail info")

	getExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")

	approveExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Colony Executor Id")
	approveExecutorCmd.MarkFlagRequired("name")

	rejectExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor Id")
	rejectExecutorCmd.MarkFlagRequired("executorid")
}

var executorCmd = &cobra.Command{
	Use:   "executors",
	Short: "Manage executors",
	Long:  "Manage executors",
}

var addExecutorCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new Executor",
	Long:  "Add a new Executor",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if len(ExecutorID) != 64 {
			CheckError(errors.New("Invalid Executor Id length"))
		}

		if ExecutorType == "" {
			CheckError(errors.New("Invalid Executor type"))
		}

		if os.Getenv("HOSTNAME") != "" {
			ExecutorName += "."
			ExecutorName += os.Getenv("HOSTNAME")
		}

		var executor *core.Executor
		if SpecFile != "" {
			jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
			CheckError(err)
			executor, err = core.ConvertJSONToExecutor(string(jsonSpecBytes))
			CheckError(err)
		} else {
			if TargetExecutorName == "" {
				CheckError(errors.New("ExecutorName must be specified if omitting spec file"))
			}
			if TargetExecutorType == "" {
				CheckError(errors.New("ExecutorType must be specified if omitting spec file"))
			}
			executor = &core.Executor{}
		}

		if TargetExecutorName != "" {
			executor.Name = TargetExecutorName
		}

		if TargetExecutorType != "" {
			executor.Type = TargetExecutorType
		}

		executor.SetID(ExecutorID)
		executor.SetColonyName(ColonyName)

		if ColonyPrvKey == "" {
			CheckError(errors.New("ERROR:" + ColonyPrvKey))
		}

		addedExecutor, err := client.AddExecutor(executor, ColonyPrvKey)
		CheckError(err)

		if Approve {
			log.WithFields(log.Fields{"ExecutorName": executor.Name}).Info("Approving Executor")
			err = client.ApproveExecutor(ColonyName, executor.Name, ColonyPrvKey)
			CheckError(err)
		}

		log.WithFields(log.Fields{
			"ExecutorName": executor.Name,
			"ExecutorType": executor.Type,
			"ExecutorID":   addedExecutor.ID,
			"ColonyName":   ColonyName}).
			Info("Executor added")
	},
}

var removeExecutorCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an Executor",
	Long:  "Remove an Executor",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if TargetExecutorID != "" {
			err := client.DeleteExecutor(TargetExecutorID, ColonyPrvKey)
			CheckError(err)
		} else {
			removeExecutorFromTmp(client)
		}

		log.WithFields(log.Fields{"TargetExecutorID": TargetExecutorID, "ColonyID": ColonyID}).Info("Executor removed")
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
		[]string{"ColonyName", executor.ColonyName},
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

	fmt.Println()
	fmt.Println("Location:")

	locationData := [][]string{
		[]string{"Longitude", fmt.Sprintf("%f", executor.Location.Long)},
		[]string{"Latitude", fmt.Sprintf("%f", executor.Location.Lat)},
		[]string{"Description", executor.Location.Description},
	}

	locationTable := tablewriter.NewWriter(os.Stdout)
	for _, v := range locationData {
		locationTable.Append(v)
	}
	locationTable.SetAlignment(tablewriter.ALIGN_LEFT)
	locationTable.Render()

	fmt.Println()
	fmt.Println("Hardware:")

	hwData := [][]string{
		[]string{"Model", executor.Capabilities.Hardware.Model},
		[]string{"CPU", executor.Capabilities.Hardware.CPU},
		[]string{"Nodes", strconv.Itoa(executor.Capabilities.Hardware.Nodes)},
		[]string{"Memory", executor.Capabilities.Hardware.Memory},
		[]string{"Storage", executor.Capabilities.Hardware.Storage},
		[]string{"GPU", executor.Capabilities.Hardware.GPU.Name},
		[]string{"GPUMem", executor.Capabilities.Hardware.GPU.Memory},
		[]string{"GPUs", strconv.Itoa(executor.Capabilities.Hardware.GPU.Count)},
		[]string{"GPUs/Node", strconv.Itoa(executor.Capabilities.Hardware.GPU.NodeCount)},
	}

	hwTable := tablewriter.NewWriter(os.Stdout)
	for _, v := range hwData {
		hwTable.Append(v)
	}
	hwTable.SetAlignment(tablewriter.ALIGN_LEFT)
	hwTable.Render()

	fmt.Println()
	fmt.Println("Software:")

	swData := [][]string{
		[]string{"Name", executor.Capabilities.Software.Name},
		[]string{"Type", executor.Capabilities.Software.Type},
		[]string{"Version", executor.Capabilities.Software.Version},
	}

	swTable := tablewriter.NewWriter(os.Stdout)
	for _, v := range swData {
		swTable.Append(v)
	}
	swTable.SetAlignment(tablewriter.ALIGN_LEFT)
	swTable.Render()

	functions, err := client.GetFunctionsByExecutorID(executor.ID, PrvKey)
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
	Short: "List all Executors",
	Long:  "List all Executors",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		executorsFromServer, err := client.GetExecutors(ColonyName, PrvKey)
		CheckError(err)

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
				data = append(data, []string{executor.Name, executor.Type, executor.Location.Description})
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Type", "Location"})

			for _, v := range data {
				table.Append(v)
			}

			table.Render()

		}
	},
}

var getExecutorCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about an Executor",
	Long:  "Get info about an Executor",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if TargetExecutorName == "" {
			CheckError(errors.New("Executor name not specified"))
		}

		executorFromServer, err := client.GetExecutor(ColonyName, TargetExecutorName, PrvKey)
		CheckError(err)

		printExecutor(client, executorFromServer)
	},
}

var approveExecutorCmd = &cobra.Command{
	Use:   "approve",
	Short: "Approve an Executor",
	Long:  "Approve an Executor",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if TargetExecutorName == "" {
			CheckError(errors.New("Executor name must be specified"))
		}

		err := client.ApproveExecutor(ColonyName, TargetExecutorName, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ExecutorName": TargetExecutorName, "ColonyName": ColonyName}).Info("Executor approved")
	},
}

var rejectExecutorCmd = &cobra.Command{
	Use:   "reject",
	Short: "Reject an Executor",
	Long:  "Reject an Executor",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if TargetExecutorName == "" {
			CheckError(errors.New("Executor name must be specified"))
		}

		err := client.RejectExecutor(ColonyName, TargetExecutorName, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ExecutorName": TargetExecutorName, "ColonyName": ColonyName}).Info("Executor reject")
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
