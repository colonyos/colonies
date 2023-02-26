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
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	colonyCmd.AddCommand(addColonyCmd)
	colonyCmd.AddCommand(removeColonyCmd)
	colonyCmd.AddCommand(lsColoniesCmd)
	colonyCmd.AddCommand(colonyStatsCmd)
	rootCmd.AddCommand(colonyCmd)

	colonyCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", DefaultServerHost, "Server host")
	colonyCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addColonyCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	addColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	addColonyCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	addColonyCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony")
	addColonyCmd.MarkFlagRequired("spec")

	removeColonyCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	removeColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	removeColonyCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	removeColonyCmd.MarkFlagRequired("colonyid")

	lsColoniesCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	lsColoniesCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	lsColoniesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	colonyStatsCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	colonyStatsCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	colonyStatsCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
}

var colonyCmd = &cobra.Command{
	Use:   "colony",
	Short: "Manage colonies",
	Long:  "Manage colonies",
}

var addColonyCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new colony",
	Long:  "Add a new colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		colony, err := core.ConvertJSONToColony(string(jsonSpecBytes))
		CheckError(err)

		crypto := crypto.CreateCrypto()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		var prvKey string
		if ColonyPrvKey != "" {
			prvKey = ColonyPrvKey
			if len(prvKey) != 64 {
				CheckError(errors.New("Invalid private key length"))
			}
		} else {
			prvKey, err = crypto.GeneratePrivateKey()
			CheckError(err)
		}

		colonyID, err := crypto.GenerateID(prvKey)
		CheckError(err)
		colony.SetID(colonyID)

		if ServerID == "" {
			ServerID = os.Getenv("COLONIES_SERVER_ID")
		}
		if ServerID == "" {
			CheckError(errors.New("Unknown Server Id"))
		}

		if ServerPrvKey == "" {
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		addedColony, err := client.AddColony(colony, ServerPrvKey)
		CheckError(err)

		err = keychain.AddPrvKey(colonyID, prvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ColonyID": addedColony.ID}).Info("Colony added")
	},
}

var removeColonyCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a colony",
	Long:  "Remove a colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ServerID == "" {
			ServerID = os.Getenv("COLONIES_SERVER_ID")
		}
		if ServerID == "" {
			CheckError(errors.New("Unknown Server Id"))
		}

		if ServerPrvKey == "" {
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		err = client.DeleteColony(ColonyID, ServerPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ColonyID": ColonyID}).Info("Colony removed")
	},
}

var lsColoniesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all colonies",
	Long:  "List all colonies",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ServerID == "" {
			ServerID = os.Getenv("COLONIES_SERVER_ID")
		}
		if ServerID == "" {
			CheckError(errors.New("Unknown Server Id"))
		}

		if ServerPrvKey == "" {
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		coloniesFromServer, err := client.GetColonies(ServerPrvKey)
		CheckError(err)

		if len(coloniesFromServer) == 0 {
			log.Warning("No colonies found")
			os.Exit(0)
		}

		if JSON {
			jsonString, err := core.ConvertColonyArrayToJSON(coloniesFromServer)
			CheckError(err)
			fmt.Println(jsonString)
			os.Exit(0)
		}

		var data [][]string
		for _, colony := range coloniesFromServer {
			data = append(data, []string{colony.ID, colony.Name})
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name"})

		for _, v := range data {
			table.Append(v)
		}

		table.Render()
	},
}

var colonyStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics about a colony",
	Long:  "Show statistics about a colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONIES_COLONY_ID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if ExecutorPrvKey == "" {
			if ExecutorID == "" {
				ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
			}
			ExecutorPrvKey, _ = keychain.GetPrvKey(ExecutorID)
		}

		if ExecutorPrvKey == "" {
			if ExecutorID == "" {
				ExecutorID = os.Getenv("COLONIES_EXECUTOR_ID")
			}
			ExecutorPrvKey, _ = keychain.GetPrvKey(ExecutorID)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		stat, err := client.ColonyStatistics(ColonyID, ExecutorPrvKey)
		CheckError(err)

		fmt.Println("Process statistics:")
		specData := [][]string{
			[]string{"Executors", strconv.Itoa(stat.Executors)},
			[]string{"Waiting processes", strconv.Itoa(stat.WaitingProcesses)},
			[]string{"Running processes", strconv.Itoa(stat.RunningProcesses)},
			[]string{"Successful processes", strconv.Itoa(stat.SuccessfulProcesses)},
			[]string{"Failed processes", strconv.Itoa(stat.FailedProcesses)},
			[]string{"Waiting workflows", strconv.Itoa(stat.WaitingWorkflows)},
			[]string{"Running workflows ", strconv.Itoa(stat.RunningWorkflows)},
			[]string{"Successful workflows", strconv.Itoa(stat.SuccessfulWorkflows)},
			[]string{"Failed workflows", strconv.Itoa(stat.FailedWorkflows)},
		}
		specTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range specData {
			specTable.Append(v)
		}
		specTable.SetAlignment(tablewriter.ALIGN_LEFT)
		specTable.Render()
	},
}
