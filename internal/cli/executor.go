package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/colonyos/colonies/pkg/core"
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
	executorCmd.PersistentFlags().StringVarP(&ColonyName, "colonyid", "", "", "Colony Id")
	executorCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	executorCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addExecutorCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of an executor")
	addExecutorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor ID")
	addExecutorCmd.MarkFlagRequired("executorid")
	addExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")
	addExecutorCmd.Flags().StringVarP(&TargetExecutorType, "type", "", "", "Executor type")
	addExecutorCmd.Flags().BoolVarP(&Approve, "approve", "", false, "Also, approve the Executor")

	removeExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor Id")

	lsExecutorsCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	lsExecutorsCmd.Flags().BoolVarP(&Full, "full", "", false, "Print detail info")

	getExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")

	approveExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Colony Executor Id")
	approveExecutorCmd.MarkFlagRequired("name")

	rejectExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor Id")
	rejectExecutorCmd.MarkFlagRequired("executorid")
}

var executorCmd = &cobra.Command{
	Use:   "executor",
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

		if TargetExecutorName == "" {
			CheckError(errors.New("Executor name must be specified"))
		}

		err := client.RemoveExecutor(ColonyName, TargetExecutorName, ColonyPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ExecutorName": TargetExecutorName, "ColonyName": ColonyName}).Info("Executor removed")
	},
}

var lsExecutorsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all Executors",
	Long:  "List all Executors",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		executorsFromServer, err := client.GetExecutors(ColonyName, PrvKey)
		CheckError(err)

		if len(executorsFromServer) == 0 {
			log.Info("No Executors found")
			os.Exit(0)
		}

		if Full {
			if JSON {
				jsonString, err := core.ConvertExecutorArrayToJSON(executorsFromServer)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			for counter, executor := range executorsFromServer {
				printExecutorTable(client, executor)

				if counter != len(executorsFromServer)-1 {
					fmt.Println()
					fmt.Println("==============================================================================================")
					fmt.Println()
				} else {
				}
			}
		} else {
			printExecutorsTable(executorsFromServer)
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

		printExecutorTable(client, executorFromServer)
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
