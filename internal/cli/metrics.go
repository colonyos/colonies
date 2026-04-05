package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	metricsCmd.AddCommand(lsMetricsCmd)
	metricsCmd.AddCommand(getMetricCmd)
	metricsCmd.AddCommand(historyMetricsCmd)
	metricsCmd.AddCommand(rmMetricCmd)
	rootCmd.AddCommand(metricsCmd)

	metricsCmd.PersistentFlags().StringVarP(&ColonyName, "colonyname", "", "", "Colony name")
	metricsCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	metricsCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	lsMetricsCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")
	lsMetricsCmd.Flags().BoolVarP(&JSON, "json", "", false, "Output as JSON")

	getMetricCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")
	getMetricCmd.MarkFlagRequired("name")
	getMetricCmd.Flags().StringVarP(&Key, "key", "", "", "Metric key")
	getMetricCmd.MarkFlagRequired("key")
	getMetricCmd.Flags().BoolVarP(&JSON, "json", "", false, "Output as JSON")

	historyMetricsCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")
	historyMetricsCmd.MarkFlagRequired("name")
	historyMetricsCmd.Flags().StringVarP(&Key, "key", "", "", "Metric key")
	historyMetricsCmd.MarkFlagRequired("key")
	historyMetricsCmd.Flags().StringVarP(&MetricPeriod, "period", "", "", "Period: day, week, or month")
	historyMetricsCmd.MarkFlagRequired("period")
	historyMetricsCmd.Flags().StringVarP(&FromDate, "from", "", "", "Start date (YYYY-MM-DD)")
	historyMetricsCmd.Flags().StringVarP(&ToDate, "to", "", "", "End date (YYYY-MM-DD)")
	historyMetricsCmd.Flags().BoolVarP(&JSON, "json", "", false, "Output as JSON")

	rmMetricCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")
	rmMetricCmd.MarkFlagRequired("name")
	rmMetricCmd.Flags().StringVarP(&Key, "key", "", "", "Metric key (omit to remove all)")
}

func parsePeriod(s string) (int, error) {
	switch s {
	case "day":
		return core.PERIOD_DAY, nil
	case "week":
		return core.PERIOD_WEEK, nil
	case "month":
		return core.PERIOD_MONTH, nil
	default:
		return 0, errors.New("Invalid period, must be: day, week, or month")
	}
}

func defaultFrom(period int) time.Time {
	now := time.Now().UTC()
	switch period {
	case core.PERIOD_DAY:
		return now.AddDate(0, 0, -7)
	case core.PERIOD_WEEK:
		return now.AddDate(0, 0, -28)
	case core.PERIOD_MONTH:
		return now.AddDate(0, -6, 0)
	default:
		return now.AddDate(0, 0, -7)
	}
}

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Manage executor metrics",
	Long:  "Manage executor metrics",
}

var lsMetricsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List metrics",
	Long:  "List metrics for an executor, or list executors with metric counts",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if TargetExecutorName != "" {
			metrics, err := client.GetAllMetrics(ColonyName, TargetExecutorName, PrvKey)
			CheckError(err)

			if len(metrics) == 0 {
				log.Info("No metrics found")
				os.Exit(0)
			}

			if JSON {
				jsonString, err := core.ConvertMetricArrayToJSON(metrics)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			printMetricKeysTable(metrics)
		} else {
			executors, err := client.GetExecutors(ColonyName, PrvKey)
			CheckError(err)

			if len(executors) == 0 {
				log.Info("No executors found")
				os.Exit(0)
			}

			var results []executorMetricCount
			for _, executor := range executors {
				metrics, err := client.GetMetrics(ColonyName, executor.Name, PrvKey)
				CheckError(err)
				results = append(results, executorMetricCount{
					Name:  executor.Name,
					Count: len(metrics),
				})
			}

			if JSON {
				type jsonResult struct {
					Executor string `json:"executor"`
					Metrics  int    `json:"metrics"`
				}
				var jsonResults []jsonResult
				for _, r := range results {
					jsonResults = append(jsonResults, jsonResult{
						Executor: r.Name,
						Metrics:  r.Count,
					})
				}
				printMetricsJSON(jsonResults)
				os.Exit(0)
			}

			printExecutorMetricCountTable(results)
		}
	},
}

func printMetricsJSON(v interface{}) {
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	CheckError(err)
	fmt.Println(string(jsonBytes))
}

var getMetricCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a metric value",
	Long:  "Get a single metric value for an executor",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		metric, err := client.GetMetric(ColonyName, TargetExecutorName, Key, PrvKey)
		CheckError(err)

		if JSON {
			jsonString, err := metric.ToJSON()
			CheckError(err)
			fmt.Println(jsonString)
			os.Exit(0)
		}

		printMetricDetailTable(metric)
	},
}

var historyMetricsCmd = &cobra.Command{
	Use:   "history",
	Short: "Show metric history",
	Long:  "Show time-series history for a metric",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		period, err := parsePeriod(MetricPeriod)
		CheckError(err)

		var from time.Time
		if FromDate != "" {
			from, err = time.Parse("2006-01-02", FromDate)
			CheckError(err)
		} else {
			from = defaultFrom(period)
		}

		var to time.Time
		if ToDate != "" {
			to, err = time.Parse("2006-01-02", ToDate)
			CheckError(err)
			// Set to end of day
			to = to.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		} else {
			to = time.Now().UTC()
		}

		metrics, err := client.GetMetricHistory(ColonyName, TargetExecutorName, Key, period, from, to, PrvKey)
		CheckError(err)

		if len(metrics) == 0 {
			log.Info("No metric history found")
			os.Exit(0)
		}

		if JSON {
			jsonString, err := core.ConvertMetricArrayToJSON(metrics)
			CheckError(err)
			fmt.Println(jsonString)
			os.Exit(0)
		}

		printMetricHistoryTable(metrics)
	},
}

var rmMetricCmd = &cobra.Command{
	Use:   "rm",
	Short: "Remove metrics",
	Long:  "Remove metrics for an executor",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if Key != "" {
			err := client.RemoveMetric(ColonyName, TargetExecutorName, Key, PrvKey)
			CheckError(err)
			log.WithFields(log.Fields{"ExecutorName": TargetExecutorName, "Key": Key}).Info("Metric removed")
		} else {
			err := client.RemoveAllMetrics(ColonyName, TargetExecutorName, PrvKey)
			CheckError(err)
			log.WithFields(log.Fields{"ExecutorName": TargetExecutorName}).Info("All metrics removed")
		}
	},
}
