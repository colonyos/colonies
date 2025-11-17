package cli

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	cronCmd.AddCommand(addCronCmd)
	cronCmd.AddCommand(delCronCmd)
	cronCmd.AddCommand(getCronCmd)
	cronCmd.AddCommand(getCronsCmd)
	cronCmd.AddCommand(runCronCmd)
	rootCmd.AddCommand(cronCmd)

	cronCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	cronCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addCronCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony workflow")
	addCronCmd.MarkFlagRequired("spec")
	addCronCmd.Flags().StringVarP(&CronName, "name", "", "", "Cron name")
	addCronCmd.MarkFlagRequired("name")
	addCronCmd.Flags().StringVarP(&CronExpr, "cron", "", "", "Cron expression")
	addCronCmd.Flags().IntVarP(&CronInterval, "interval", "", -1, "Interval in seconds")
	addCronCmd.Flags().BoolVarP(&CronRandom, "random", "", false, "Schedule a random cron, interval must be specified")
	addCronCmd.Flags().BoolVarP(&WaitForPrevProcessGraph, "waitprevious", "", false, "Wait for previous processgrah to finish bore schedule a new workflow")

	delCronCmd.Flags().StringVarP(&CronID, "cronid", "", "", "Cron Id")
	delCronCmd.MarkFlagRequired("cronid")

	getCronCmd.Flags().StringVarP(&CronID, "cronid", "", "", "Cron Id")
	getCronCmd.MarkFlagRequired("cronid")

	getCronsCmd.Flags().IntVarP(&Count, "count", "", DefaultCount, "Number of crons to list")

	runCronCmd.Flags().StringVarP(&CronID, "cronid", "", "", "Cron Id")
	runCronCmd.MarkFlagRequired("cronid")
}

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Manage cron",
	Long:  "Manage cron",
}

var addCronCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a cron",
	Long:  "Add a cron",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		jsonStr := "{\"functionspecs\":" + string(jsonSpecBytes) + "}"
		workflowSpec, err := core.ConvertJSONToWorkflowSpec(jsonStr)
		CheckError(err)

		if workflowSpec.ColonyName == "" {
			workflowSpec.ColonyName = ColonyName
		}

		workflowSpecJSON, err := workflowSpec.ToJSON()
		CheckError(err)

		if workflowSpec.ColonyName == "" {
			workflowSpec.ColonyName = ColonyName
		}

		if CronName == "" {
			CheckError(errors.New("Cron name not specified"))
		}

		if CronInterval == -1 && CronExpr == "-1" {
			CheckError(errors.New("Cron expression or interval must be specified"))
		}

		cron := core.CreateCron(ColonyName, CronName, CronExpr, CronInterval, CronRandom, workflowSpecJSON)

		if WaitForPrevProcessGraph {
			log.Info("Waiting for previous processgraph to finish")
			cron.WaitForPrevProcessGraph = true
		} else {
			log.Info("Will not wait for previous processgraph to finish")
		}

		addedCron, err := client.AddCron(cron, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"CronID": addedCron.ID}).Info("Cron added")
	},
}

var delCronCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a cron",
	Long:  "Remove a cron",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if CronID == "" {
			CheckError(errors.New("Cron Id not specified"))
		}

		err := client.RemoveCron(CronID, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"CronId": CronID}).Info("Removing cron")
	},
}

var getCronCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a cron",
	Long:  "Get info about a cron",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if CronID == "" {
			CheckError(errors.New("Cron Id not specified"))
		}

		cron, err := client.GetCron(CronID, PrvKey)
		if cron == nil {
			log.WithFields(log.Fields{"CronId": CronID}).Error("Cron not found")
			os.Exit(0)
		}

		printCronTable(cron)

		workflowSpec, err := core.ConvertJSONToWorkflowSpec(cron.WorkflowSpec)
		CheckError(err)
		for _, funcSpec := range workflowSpec.FunctionSpecs {
			printFunctionSpecTable(&funcSpec)
		}
	},
}

var getCronsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all crons",
	Long:  "List all crons",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		crons, err := client.GetCrons(ColonyName, Count, PrvKey)
		CheckError(err)
		if crons == nil || len(crons) == 0 {
			log.WithFields(log.Fields{"ColonyName": ColonyName}).Info("No crons found")
			os.Exit(0)
		}

		printCronsTable(crons)
	},
}

var runCronCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a cron now",
	Long:  "Run a cron now",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if CronID == "" {
			CheckError(errors.New("Cron Id not specified"))
		}

		_, err := client.RunCron(CronID, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"CronID": CronID}).Info("Running cron")
	},
}
