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
	"github.com/colonyos/colonies/pkg/server"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	colonyCmd.AddCommand(registerColonyCmd)
	colonyCmd.AddCommand(unregisterColonyCmd)
	colonyCmd.AddCommand(lsColoniesCmd)
	colonyCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(colonyCmd)

	colonyCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", DefaultServerHost, "Server host")
	colonyCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", DefaultServerPort, "Server HTTP port")

	registerColonyCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	registerColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	registerColonyCmd.Flags().StringVarP(&ColonyPrvKey, "colonyprvkey", "", "", "Colony private key")
	registerColonyCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony")
	registerColonyCmd.MarkFlagRequired("spec")

	unregisterColonyCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	unregisterColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	unregisterColonyCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	unregisterColonyCmd.MarkFlagRequired("colonyid")

	lsColoniesCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	lsColoniesCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	lsColoniesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")

	statusCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	statusCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	statusCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
}

var colonyCmd = &cobra.Command{
	Use:   "colony",
	Short: "Manage Colonies",
	Long:  "Manage Colonies",
}

var registerColonyCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new Colony",
	Long:  "Register a new Colony",
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
				CheckError(errors.New("invalid private key length"))
			}
		} else {
			prvKey, err = crypto.GeneratePrivateKey()
			CheckError(err)
		}

		colonyID, err := crypto.GenerateID(prvKey)
		CheckError(err)
		colony.SetID(colonyID)

		if ServerID == "" {
			ServerID = os.Getenv("COLONIES_SERVERID")
		}
		if ServerID == "" {
			CheckError(errors.New("Unknown Server Id"))
		}

		if ServerPrvKey == "" {
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		addedColony, err := client.AddColony(colony, ServerPrvKey)
		CheckError(err)

		err = keychain.AddPrvKey(colonyID, prvKey)
		CheckError(err)

		log.WithFields(log.Fields{"colonyID": addedColony.ID}).Info("Colony registered")
	},
}

var unregisterColonyCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister a Colony",
	Long:  "Unregister a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ServerID == "" {
			ServerID = os.Getenv("COLONIES_SERVERID")
		}
		if ServerID == "" {
			CheckError(errors.New("Unknown Server Id"))
		}

		if ServerPrvKey == "" {
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		err = client.DeleteColony(ColonyID, ServerPrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"colonyID": ColonyID}).Info("Colony unregistered")
	},
}

var lsColoniesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all Colonies",
	Long:  "List all Colonies",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ServerID == "" {
			ServerID = os.Getenv("COLONIES_SERVERID")
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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status about a Colony",
	Long:  "Show status about a Colony",
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

		if RuntimePrvKey == "" {
			if RuntimeID == "" {
				RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
			}
			RuntimePrvKey, _ = keychain.GetPrvKey(RuntimeID)
		}

		if RuntimePrvKey == "" {
			if RuntimeID == "" {
				RuntimeID = os.Getenv("COLONIES_RUNTIMEID")
			}
			RuntimePrvKey, _ = keychain.GetPrvKey(RuntimeID)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		runtimesFromServer, err := client.GetRuntimes(ColonyID, RuntimePrvKey)
		prvKey := RuntimePrvKey
		if err != nil {
			if ColonyPrvKey == "" {
				if ColonyID == "" {
					ColonyID = os.Getenv("COLONIES_COLONYID")
				}
				ColonyPrvKey, err = keychain.GetPrvKey(ColonyID)
				CheckError(err)
			}
			runtimesFromServer, err = client.GetRuntimes(ColonyID, ColonyPrvKey)
			CheckError(err)
			prvKey = ColonyPrvKey
		}

		waitingProcesses3600, err := client.GetProcessHistForColony(core.WAITING, ColonyID, 3600, prvKey)
		CheckError(err)
		runningProcesses3600, err := client.GetProcessHistForColony(core.RUNNING, ColonyID, 3600, prvKey)
		CheckError(err)
		successfulProcesses3600, err := client.GetProcessHistForColony(core.SUCCESS, ColonyID, 3600, prvKey)
		CheckError(err)
		failedProcesses3600, err := client.GetProcessHistForColony(core.FAILED, ColonyID, 3600, prvKey)
		CheckError(err)

		var allProcesses3600 []*core.Process
		allProcesses3600 = append(allProcesses3600, waitingProcesses3600...)
		allProcesses3600 = append(allProcesses3600, runningProcesses3600...)
		allProcesses3600 = append(allProcesses3600, successfulProcesses3600...)
		allProcesses3600 = append(allProcesses3600, failedProcesses3600...)
		retries3600 := utils.CalcRetries(allProcesses3600)

		processes, err := client.GetRunningProcesses(ColonyID, server.MAX_COUNT-1, prvKey)
		CheckError(err)

		stat, err := client.GetProcessStat(ColonyID, RuntimePrvKey)
		CheckError(err)

		fmt.Println("Process statistics:")
		specData := [][]string{
			[]string{"Waiting processes", strconv.Itoa(stat.Waiting)},
			[]string{"Running processes ", strconv.Itoa(stat.Running)},
			[]string{"Successful processes", strconv.Itoa(stat.Success)},
			[]string{"Runtimes", strconv.Itoa(len(runtimesFromServer))},
			[]string{"Failed", strconv.Itoa(stat.Failed)},
			[]string{"Retries (1 hour)", strconv.Itoa(retries3600)},
			[]string{"Utilization (1 hour)", fmt.Sprintf("%f", utils.CalcUtilization(successfulProcesses3600)) + "%"},
			[]string{"Avg waiting time (1 h)", fmt.Sprintf("%f", utils.CalcAvgWaitingTime(successfulProcesses3600)) + "s"},
			[]string{"Avg processing time (1 h)", fmt.Sprintf("%f", utils.CalcAvgProcessingTime(successfulProcesses3600)) + "s"},
		}
		specTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range specData {
			specTable.Append(v)
		}
		specTable.SetAlignment(tablewriter.ALIGN_LEFT)
		specTable.Render()

		runningRuntimes := make(map[string]bool)
		for _, p := range processes {
			runningRuntimes[p.AssignedRuntimeID] = true
		}

		cores := 0
		for _, runtime := range runtimesFromServer {
			cores += runtime.Cores
		}

		mem := 0
		for _, runtime := range runtimesFromServer {
			mem += runtime.Mem
		}

		gpus := 0
		for _, runtime := range runtimesFromServer {
			gpus += runtime.GPUs
		}
	},
}
