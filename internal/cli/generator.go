package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/server"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	generatorCmd.AddCommand(addGeneratorCmd)
	generatorCmd.AddCommand(packGeneratorCmd)
	generatorCmd.AddCommand(delGeneratorCmd)
	generatorCmd.AddCommand(getGeneratorCmd)
	generatorCmd.AddCommand(getGeneratorsCmd)
	rootCmd.AddCommand(generatorCmd)

	generatorCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	generatorCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addGeneratorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	addGeneratorCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	addGeneratorCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony workflow")
	addGeneratorCmd.MarkFlagRequired("spec")
	addGeneratorCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	addGeneratorCmd.Flags().StringVarP(&GeneratorName, "name", "", "", "Generator name")
	addGeneratorCmd.MarkFlagRequired("name")
	addGeneratorCmd.Flags().IntVarP(&GeneratorTrigger, "trigger", "", -1, "Trigger")
	addGeneratorCmd.MarkFlagRequired("trigger")
	addGeneratorCmd.Flags().IntVarP(&GeneratorTimeout, "timeout", "", -1, "Timeout")

	packGeneratorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	packGeneratorCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	packGeneratorCmd.Flags().StringVarP(&GeneratorID, "generatorid", "", "", "Generator Id")
	packGeneratorCmd.MarkFlagRequired("generatorid")
	packGeneratorCmd.Flags().StringVarP(&Arg, "arg", "", "", "Arg to pack to generator")
	packGeneratorCmd.MarkFlagRequired("arg")

	delGeneratorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	delGeneratorCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	delGeneratorCmd.Flags().StringVarP(&GeneratorID, "generatorid", "", "", "Generator Id")
	delGeneratorCmd.MarkFlagRequired("generatorid")

	getGeneratorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	getGeneratorCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	getGeneratorCmd.Flags().StringVarP(&GeneratorID, "generatorid", "", "", "Generator Id")
	getGeneratorCmd.MarkFlagRequired("generatorid")

	getGeneratorsCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	getGeneratorsCmd.Flags().StringVarP(&ExecutorPrvKey, "executorprvkey", "", "", "Executor private key")
	getGeneratorsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	getGeneratorsCmd.Flags().IntVarP(&Count, "count", "", server.MAX_COUNT, "Number of generators to list")
}

var generatorCmd = &cobra.Command{
	Use:   "generator",
	Short: "Manage generators",
	Long:  "Manage generators",
}

var addGeneratorCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a generator",
	Long:  "Add a generator",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		jsonStr := "{\"functionspecs\":" + string(jsonSpecBytes) + "}"
		workflowSpec, err := core.ConvertJSONToWorkflowSpec(jsonStr)
		CheckError(err)

		if workflowSpec.ColonyID == "" {
			workflowSpec.ColonyID = ColonyID
		}

		workflowSpecJSON, err := workflowSpec.ToJSON()
		CheckError(err)

		if GeneratorName == "" {
			CheckError(errors.New("Generator name not specified"))
		}

		if GeneratorTrigger == -1 {
			CheckError(errors.New("Generator trigger not specified"))
		}

		generator := core.CreateGenerator(ColonyID, GeneratorName, workflowSpecJSON, GeneratorTrigger, GeneratorTimeout)
		addedGenerator, err := client.AddGenerator(generator, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"GeneratorID": addedGenerator.ID, "GeneratorName": GeneratorName, "Trigger": GeneratorTrigger, "Timeout": GeneratorTimeout}).Info("Generator added")
	},
}

var packGeneratorCmd = &cobra.Command{
	Use:   "pack",
	Short: "Pack arg to a generator",
	Long:  "Pack arg to a generator",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if GeneratorID == "" {
			CheckError(errors.New("Generator Id not specified"))
		}

		err := client.PackGenerator(GeneratorID, Arg, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"GeneratorID": GeneratorID, "Arg": Arg}).Info("Packing arg to generator")
	},
}

var delGeneratorCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a generator",
	Long:  "Delete a generator",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if GeneratorID == "" {
			CheckError(errors.New("Generator Id not specified"))
		}

		err := client.DeleteGenerator(GeneratorID, ExecutorPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"GeneratorID": GeneratorID}).Info("Deleting generator")
	},
}

var getGeneratorCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a generator",
	Long:  "Get info about a generator",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if GeneratorID == "" {
			CheckError(errors.New("Generator Id not specified"))
		}

		generator, err := client.GetGenerator(GeneratorID, ExecutorPrvKey)
		if generator == nil {
			log.WithFields(log.Fields{"GeneratorId": GeneratorID}).Error("Generator not found")
			os.Exit(0)
		}

		fmt.Println("Generator:")
		generatorData := [][]string{
			[]string{"Id", generator.ID},
			[]string{"Name", generator.Name},
			[]string{"Trigger", strconv.Itoa(generator.Trigger)},
			[]string{"Timeout", strconv.Itoa(generator.Timeout)},
			[]string{"Lastrun", generator.LastRun.Format(TimeLayout)},
			[]string{"CheckerPeriod", strconv.Itoa(generator.CheckerPeriod)},
			[]string{"QueueSize", strconv.Itoa(generator.QueueSize)},
		}
		generatorTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range generatorData {
			generatorTable.Append(v)
		}
		generatorTable.SetAlignment(tablewriter.ALIGN_LEFT)
		generatorTable.SetAutoWrapText(false)
		generatorTable.Render()

		fmt.Println()
		fmt.Println("WorkflowSpec:")
		workflowSpec, err := core.ConvertJSONToWorkflowSpec(generator.WorkflowSpec)
		CheckError(err)
		for i, funcSpec := range workflowSpec.FunctionSpecs {
			fmt.Println()
			fmt.Println("FunctionSpec " + strconv.Itoa(i) + ":")
			printFunctionSpec(&funcSpec)
		}
	},
}

var getGeneratorsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all generators",
	Long:  "List all generators",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		generators, err := client.GetGenerators(ColonyID, Count, ExecutorPrvKey)
		CheckError(err)
		if generators == nil {
			log.WithFields(log.Fields{"ColonyId": ColonyID}).Info("No generators found")
			os.Exit(0)
		}

		var data [][]string
		for _, generator := range generators {
			data = append(data, []string{generator.ID, generator.Name})
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"GeneratorId", "Name"})
		for _, v := range data {
			table.Append(v)
		}
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.Render()
	},
}
