package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/colonyos/colonies/internal/table"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/muesli/termenv"
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
	getLogsCmd.Flags().BoolVarP(&Latest, "latest", "l", true, "Show latest logs (most recent)")
	getLogsCmd.Flags().BoolVarP(&First, "first", "f", false, "Show logs from the start")

	searchLogsCmd.Flags().StringVarP(&Text, "text", "t", "", "Text to search")
	searchLogsCmd.Flags().IntVarP(&Days, "days", "d", 1, "Number of days back in time to search")
	searchLogsCmd.Flags().IntVarP(&Count, "count", "", 10, "Number of search results to fetch")
	searchLogsCmd.Flags().BoolVarP(&Print, "print", "", false, "Print logs")
	searchLogsCmd.Flags().IntVarP(&SecondsBack, "seconds", "", 1, "Seconds back in time to print logs before the search result occurred")
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

// logrusPattern matches logrus-style log format: time="..." level=... msg="..."
var logrusPattern = regexp.MustCompile(`^time="([^"]+)"\s+level=(\w+)\s+msg="(.*)"$`)

// formatLogMessage parses and colorizes logrus-style log messages
// Falls back to plain output for non-matching formats
func formatLogMessage(message string, theme table.Theme) string {
	// Trim trailing newline for parsing, we'll add it back
	trimmed := strings.TrimSuffix(message, "\n")

	matches := logrusPattern.FindStringSubmatch(trimmed)
	if matches == nil {
		// Not logrus format, return as-is
		return message
	}

	timestamp := matches[1]
	level := matches[2]
	msg := matches[3]

	// Color the level based on severity
	var levelColored string
	switch strings.ToLower(level) {
	case "error", "fatal", "panic":
		levelColored = termenv.String(strings.ToUpper(level)).Foreground(theme.ColorRed).Bold().String()
	case "warn", "warning":
		levelColored = termenv.String(strings.ToUpper(level)).Foreground(theme.ColorYellow).Bold().String()
	case "info":
		levelColored = termenv.String(strings.ToUpper(level)).Foreground(theme.ColorGreen).String()
	case "debug", "trace":
		levelColored = termenv.String(strings.ToUpper(level)).Foreground(theme.ColorGray).String()
	default:
		levelColored = termenv.String(strings.ToUpper(level)).Foreground(theme.ColorBlue).String()
	}

	// Format: [timestamp] LEVEL  message
	timeColored := termenv.String(timestamp).Foreground(theme.ColorCyan).String()

	return fmt.Sprintf("[%s] %-7s %s\n", timeColored, levelColored, msg)
}

var getLogsCmd = &cobra.Command{
	Use:   "get",
	Short: "Get logs added to a process",
	Long:  "Get logs added to a process",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		theme, err := table.LoadTheme("solarized-dark")
		CheckError(err)

		// If --first is specified, disable --latest
		useLatest := Latest && !First

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
					for _, logEntry := range logs {
						fmt.Print(formatLogMessage(logEntry.Message, theme))
					}
					lastTimestamp = logs[len(logs)-1].Timestamp
				}
			}
		} else {
			var logs []*core.Log
			if TargetExecutorName == "" {
				if useLatest && Since == 0 {
					// Get latest logs (most recent count logs)
					logs, err = client.GetLogsByProcessLatest(ColonyName, ProcessID, Count, PrvKey)
				} else {
					// Get logs from start or since timestamp
					logs, err = client.GetLogsByProcessSince(ColonyName, ProcessID, Count, Since, PrvKey)
				}
			} else {
				if useLatest && Since == 0 {
					// Get latest logs (most recent count logs)
					logs, err = client.GetLogsByExecutorLatest(ColonyName, TargetExecutorName, Count, PrvKey)
				} else {
					// Get logs from start or since timestamp
					logs, err = client.GetLogsByExecutorSince(ColonyName, TargetExecutorName, Count, Since, PrvKey)
				}
			}
			CheckError(err)
			for _, logEntry := range logs {
				fmt.Print(formatLogMessage(logEntry.Message, theme))
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

		logs, err := client.SearchLogs(ColonyName, Text, Days, Count, PrvKey)
		CheckError(err)

		if len(logs) == 0 {
			log.Info("No logs found")
			os.Exit(0)
		}

		if Print {
			var timestamp int64
			timestamp = 1000000000 * int64(SecondsBack)

			log.WithFields(log.Fields{"Text": Text, "Days": Days, "Count": Count, "Seconds": SecondsBack}).Info("Searching for logs")
			fmt.Println()

			theme, err := table.LoadTheme("solarized-dark")
			CheckError(err)
			for _, log := range logs {
				fullLogs, err := client.GetLogsByProcessSince(ColonyName, log.ProcessID, Count, timestamp, PrvKey)
				CheckError(err)
				fmt.Print(termenv.String("Timestamp: ").Foreground(theme.ColorCyan))
				fmt.Println(termenv.String(time.Unix(0, log.Timestamp).Format(TimeLayout)).Foreground(theme.ColorGray))
				fmt.Print(termenv.String("ProcessID: ").Foreground(theme.ColorCyan))
				fmt.Println(termenv.String(log.ProcessID).Foreground(theme.ColorGray))
				fmt.Print(termenv.String("ExecutorName: ").Foreground(theme.ColorCyan))
				fmt.Println(termenv.String(log.ExecutorName).Foreground(theme.ColorGray))

				fmt.Println(termenv.String("=================== LOGS ====================").Foreground(theme.ColorViolet))

				for _, fullLog := range fullLogs {
					fmt.Print(fullLog.Message)
				}
				fmt.Println(termenv.String("================= END LOGS ==================").Foreground(theme.ColorViolet))
				fmt.Println()
			}
			os.Exit(0)
		}

		log.WithFields(log.Fields{"Text": Text, "Days": Days, "Count": Count}).Info("Searching for logs")
		printLogTable(logs)
	},
}
