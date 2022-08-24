package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
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

	addGeneratorCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	addGeneratorCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	addGeneratorCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony workflow")
	addGeneratorCmd.MarkFlagRequired("spec")
	addGeneratorCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	addGeneratorCmd.Flags().StringVarP(&GeneratorName, "name", "", "", "Generator name")
	addGeneratorCmd.MarkFlagRequired("name")
	addGeneratorCmd.Flags().IntVarP(&GeneratorTrigger, "trigger", "", -1, "Trigger")
	addGeneratorCmd.MarkFlagRequired("trigger")

	packGeneratorCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	packGeneratorCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	packGeneratorCmd.Flags().StringVarP(&GeneratorID, "generatorid", "", "", "Generator Id")
	packGeneratorCmd.MarkFlagRequired("generatorid")
	packGeneratorCmd.Flags().StringVarP(&Arg, "arg", "", "", "Arg to pack to generator")
	packGeneratorCmd.MarkFlagRequired("arg")

	delGeneratorCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	delGeneratorCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	delGeneratorCmd.Flags().StringVarP(&GeneratorID, "generatorid", "", "", "Generator Id")
	delGeneratorCmd.MarkFlagRequired("generatorid")

	getGeneratorCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	getGeneratorCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	getGeneratorCmd.Flags().StringVarP(&GeneratorID, "generatorid", "", "", "Generator Id")
	getGeneratorCmd.MarkFlagRequired("generatorid")

	getGeneratorsCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	getGeneratorsCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
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
	Short: "Add a generator to a Colony",
	Long:  "Add a generator to a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		jsonStr := "{\"processspecs\":" + string(jsonSpecBytes) + "}"
		workflowSpec, err := core.ConvertJSONToWorkflowSpec(jsonStr)
		CheckError(err)

		if workflowSpec.ColonyID == "" {
			if ColonyID == "" {
				ColonyID = os.Getenv("COLONIES_COLONYID")
			}
			if ColonyID == "" {
				CheckError(errors.New("Unknown Colony Id, please set COLONYID env variable or specify ColonyID in JSON file"))
			}

			workflowSpec.ColonyID = ColonyID
		}

		workflowSpecJSON, err := workflowSpec.ToJSON()
		CheckError(err)

		if workflowSpec.ColonyID == "" {
			if ColonyID == "" {
				ColonyID = os.Getenv("COLONIES_COLONYID")
			}
			if ColonyID == "" {
				CheckError(errors.New("Unknown Colony Id, please set COLONYID env variable or specify ColonyID in JSON file"))
			}

			workflowSpec.ColonyID = ColonyID
		}

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		if GeneratorName == "" {
			CheckError(errors.New("Generator name not specified"))
		}

		if GeneratorTimeout == -1 {
			CheckError(errors.New("Generator timeout not specified"))
		}

		if GeneratorTrigger == -1 {
			CheckError(errors.New("Generator trigger not specified"))
		}

		generator := core.CreateGenerator(ColonyID, GeneratorName, workflowSpecJSON, GeneratorTrigger)
		addedGenerator, err := client.AddGenerator(generator, RuntimePrvKey)

		log.WithFields(log.Fields{"GeneratorID": addedGenerator.ID}).Info("Generator added")
	},
}

var packGeneratorCmd = &cobra.Command{
	Use:   "pack",
	Short: "Pack arg to a generator",
	Long:  "Pack arg to a generator",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		if GeneratorID == "" {
			CheckError(errors.New("Generator Id not specified"))
		}

		err = client.PackGenerator(GeneratorID, Arg, RuntimePrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"GeneratorID": GeneratorID, "Arg": Arg}).Info("Packing arg to generator")
	},
}

var delGeneratorCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a generator",
	Long:  "Delete a generator",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		if GeneratorID == "" {
			CheckError(errors.New("Generator Id not specified"))
		}

		err = client.DeleteGenerator(GeneratorID, RuntimePrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"GeneratorID": GeneratorID}).Info("Deleting generator")
	},
}

var getGeneratorCmd = &cobra.Command{
	Use:   "get",
	Short: "Get info about a generator",
	Long:  "Get info about a generator",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		if GeneratorID == "" {
			CheckError(errors.New("Generator Id not specified"))
		}

		generator, err := client.GetGenerator(GeneratorID, RuntimePrvKey)
		if generator == nil {
			log.WithFields(log.Fields{"GeneratorId": GeneratorID}).Error("Generator not found")
			os.Exit(0)
		}

		fmt.Println("Generator:")
		generatorData := [][]string{
			[]string{"Id", generator.ID},
			[]string{"Name", generator.Name},
			[]string{"Trigger", strconv.Itoa(generator.Trigger)},
			[]string{"Lastrun", generator.LastRun.Format(TimeLayout)},
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
		for i, procesSpec := range workflowSpec.ProcessSpecs {
			fmt.Println()
			fmt.Println("ProcessSpec " + strconv.Itoa(i) + ":")
			printProcessSpec(&procesSpec)
		}
	},
}

var getGeneratorsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all generators in a colony",
	Long:  "List all generators in a colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		generators, err := client.GetGenerators(ColonyID, Count, RuntimePrvKey)
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
