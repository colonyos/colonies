package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	logCmd.AddCommand(addLogCmd)
	logCmd.AddCommand(getLogsCmd)
	rootCmd.AddCommand(logCmd)

	addLogCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	addLogCmd.MarkFlagRequired("processid")
	addLogCmd.Flags().StringVarP(&LogMsg, "msg", "m", "", "Message")
	addLogCmd.MarkFlagRequired("msg")

	getLogsCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	getLogsCmd.Flags().StringVarP(&TargetExecutorID, "executorid", "e", "", "Executor Id")
	getLogsCmd.Flags().Int64VarP(&Since, "since", "", 0, "Fetch log generated since (unix nano) time")
	getLogsCmd.Flags().IntVarP(&Count, "count", "", 100, "Number of messages to fetch")
	getLogsCmd.Flags().BoolVarP(&Follow, "follow", "", false, "Follow process")
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Manage logging",
	Long:  "Manage logging",
}

var addLogCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a log to an assigned process",
	Long:  "Add a log to an assigned process",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

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

		err = client.AddLog(ProcessID, LogMsg, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessID": ProcessID, "LogMsg": LogMsg}).Info("Adding log")
	},
}

var getLogsCmd = &cobra.Command{
	Use:   "get",
	Short: "Get logs added to a process",
	Long:  "Get logs added to a process",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

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

		fmt.Println(TargetExecutorID)

		if Follow {
			var logs []core.Log
			var lastTimestamp int64
			lastTimestamp = 0
			for {
				if TargetExecutorID == "" {
					logs, err = client.GetLogsByProcessIDSince(ProcessID, Count, lastTimestamp, ExecutorPrvKey)
				} else {
					logs, err = client.GetLogsByExecutorIDSince(TargetExecutorID, Count, lastTimestamp, ExecutorPrvKey)
				}
				CheckError(err)
				if len(logs) == 0 {
					time.Sleep(1 * time.Second)
					continue
				} else {
					for _, log := range logs {
						fmt.Print(log.Message)
					}
					lastTimestamp = logs[len(logs)-1].Timestamp
				}
			}
		} else {
			var err error
			var logs []core.Log
			if TargetExecutorID == "" {
				logs, err = client.GetLogsByProcessIDSince(ProcessID, Count, Since, ExecutorPrvKey)
			} else {
				logs, err = client.GetLogsByExecutorIDSince(ExecutorID, Count, Since, ExecutorPrvKey)
			}
			CheckError(err)
			for _, log := range logs {
				fmt.Print(log.Message)
			}
		}
	},
}
