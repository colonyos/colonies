package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	logCmd.AddCommand(addLogCmd)
	logCmd.AddCommand(getLogsCmd)
	logCmd.AddCommand(searchLogsCmd)
	rootCmd.AddCommand(logCmd)

	addLogCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	addLogCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	addLogCmd.MarkFlagRequired("processid")
	addLogCmd.Flags().StringVarP(&LogMsg, "msg", "m", "", "Message")
	addLogCmd.MarkFlagRequired("msg")

	getLogsCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	getLogsCmd.Flags().StringVarP(&TargetExecutorName, "executorname", "e", "", "Executor name")
	getLogsCmd.Flags().Int64VarP(&Since, "since", "", 0, "Fetch log generated since (unix nano) time")
	getLogsCmd.Flags().IntVarP(&Count, "count", "", 100, "Number of messages to fetch")
	getLogsCmd.Flags().BoolVarP(&Follow, "follow", "", false, "Follow process")

	searchLogsCmd.Flags().StringVarP(&Text, "text", "t", "", "Text to search")
	searchLogsCmd.Flags().IntVarP(&Days, "days", "d", 1, "Number of days back in time to search")
	searchLogsCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of search results to fetch")
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
		client := setup()

		err := client.AddLog(ProcessID, LogMsg, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessID": ProcessID, "LogMsg": LogMsg}).Info("Adding log")
	},
}

var getLogsCmd = &cobra.Command{
	Use:   "get",
	Short: "Get logs added to a process",
	Long:  "Get logs added to a process",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		var err error
		if Follow {
			var logs []*core.Log
			var lastTimestamp int64
			lastTimestamp = 0
			for {
				if TargetExecutorName == "" {
					logs, err = client.GetLogsByProcessSince(ColonyName, ProcessID, Count, lastTimestamp, PrvKey)
				} else {
					logs, err = client.GetLogsByExecutorSince(ColonyName, TargetExecutorName, Count, lastTimestamp, PrvKey)
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
			var logs []*core.Log
			if TargetExecutorName == "" {
				logs, err = client.GetLogsByProcessSince(ColonyName, ProcessID, Count, Since, PrvKey)
			} else {
				logs, err = client.GetLogsByExecutorSince(ColonyName, TargetExecutorName, Count, Since, PrvKey)
			}
			CheckError(err)
			for _, log := range logs {
				fmt.Print(log.Message)
			}
		}
	},
}

var searchLogsCmd = &cobra.Command{
	Use:   "search",
	Short: "Search logs",
	Long:  "Search logs",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if Text == "" {
			CheckError(fmt.Errorf("--text flag is required"))
		}

		log.WithFields(log.Fields{"Text": Text, "Days": Days, "Count": Count}).Info("Searching for logs")

		logs, err := client.SearchLogs(ColonyName, Text, Days, Count, PrvKey)
		CheckError(err)

		if len(logs) == 0 {
			log.Info("No logs found")
			os.Exit(0)
		}

		printLogTable(logs)
	},
}
