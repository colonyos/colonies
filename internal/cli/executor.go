package cli

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	executorCmd.AddCommand(addExecutorCmd)
	executorCmd.AddCommand(CreateExecutorCmd)
	executorCmd.AddCommand(chExecutorIDCmd)
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
	addExecutorCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	addExecutorCmd.MarkFlagRequired("executorid")
	addExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")
	addExecutorCmd.Flags().StringVarP(&TargetExecutorType, "type", "", "", "Executor type")
	addExecutorCmd.Flags().BoolVarP(&Approve, "approve", "", false, "Also, approve the Executor")

	CreateExecutorCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of an executor")
	CreateExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor name")
	CreateExecutorCmd.Flags().StringVarP(&TargetExecutorType, "type", "", "", "Executor type")
	CreateExecutorCmd.Flags().BoolVarP(&Approve, "approve", "", false, "Also, approve the Executor")
	CreateExecutorCmd.Flags().StringVarP(&PrvKeyPath, "keypath", "", "", "Path where the private key will be stored")
	CreateExecutorCmd.Flags().StringVarP(&IDPath, "idpath", "", "", "Path where the ID will be stored")

	chExecutorIDCmd.Flags().StringVarP(&ExecutorID, "executorid", "", "", "Executor Id")
	chExecutorIDCmd.MarkFlagRequired("executorid")

	removeExecutorCmd.Flags().StringVarP(&TargetExecutorName, "name", "", "", "Executor Id")

	lsExecutorsCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	lsExecutorsCmd.Flags().BoolVarP(&Full, "full", "", false, "Print detail info")
	lsExecutorsCmd.Flags().BoolVarP(&All, "all", "", false, "Show all executors including unregistered ones")
	lsExecutorsCmd.Flags().StringVarP(&TargetExecutorType, "type", "", "", "Filter by executor type")
	lsExecutorsCmd.Flags().StringVarP(&TargetLocation, "location", "", "", "Filter by node location")
	lsExecutorsCmd.Flags().StringVarP(&Filter, "filter", "f", "", "Filter by name or type containing string")

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

var CreateExecutorCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate keys and add a new Executor",
	Long:  "Generate keys and add a new Executor",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		crypto := crypto.CreateCrypto()
		prvKey, err := crypto.GeneratePrivateKey()
		CheckError(err)

		ExecutorID, err := crypto.GenerateID(prvKey)
		CheckError(err)

		// Save the private key to PrvKeyPath
		if PrvKeyPath != "" {
			err = ioutil.WriteFile(PrvKeyPath, []byte(prvKey), 0644)
			CheckError(err)
		} else {
			CheckError(errors.New("Private key path not specified"))
		}

		// Save the ID to IDPath
		if IDPath != "" {
			err = ioutil.WriteFile(IDPath, []byte(ExecutorID), 0644)
			CheckError(err)
		} else {
			CheckError(errors.New("ID path not specified"))
		}

		log.WithFields(log.Fields{"ExecutorId": ExecutorID}).Info("Generated Executor Id")

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
			"PrvKeyPath":   PrvKeyPath,
			"IDPath":       IDPath,
			"ColonyName":   ColonyName}).
			Info("Executor added")
	},
}

var chExecutorIDCmd = &cobra.Command{
	Use:   "chid",
	Short: "Change executor Id",
	Long:  "Change executor Id",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		if len(ExecutorID) != 64 {
			CheckError(errors.New("Invalid executor Id length"))
		}

		err := client.ChangeExecutorID(ColonyName, ExecutorID, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{
			"ColonyName": ColonyName,
			"ExecutorId": ExecutorID}).
			Info("Changed executor Id")
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

		// Filter by state, type and/or location
		var filteredExecutors []*core.Executor
		for _, executor := range executorsFromServer {
			// Filter out UNREGISTERED executors unless --all flag is set
			if !All && executor.State == core.UNREGISTERED {
				continue
			}

			// Filter by type
			if TargetExecutorType != "" && executor.Type != TargetExecutorType {
				continue
			}

			// Filter by location
			if TargetLocation != "" {
				if executor.Location.Description != TargetLocation {
					continue
				}
			}

			// Filter by name or type containing string
			if Filter != "" {
				filterLower := strings.ToLower(Filter)
				nameLower := strings.ToLower(executor.Name)
				typeLower := strings.ToLower(executor.Type)
				if !strings.Contains(nameLower, filterLower) && !strings.Contains(typeLower, filterLower) {
					continue
				}
			}

			filteredExecutors = append(filteredExecutors, executor)
		}

		if len(filteredExecutors) == 0 {
			log.Info("No Executors found")
			os.Exit(0)
		}

		if Full {
			if JSON {
				jsonString, err := core.ConvertExecutorArrayToJSON(filteredExecutors)
				CheckError(err)
				fmt.Println(jsonString)
				os.Exit(0)
			}

			for counter, executor := range filteredExecutors {
				printExecutorTable(client, executor)

				if counter != len(filteredExecutors)-1 {
					fmt.Println()
					fmt.Println("==============================================================================================")
					fmt.Println()
				} else {
				}
			}
		} else {
			printExecutorsTable(filteredExecutors, All)
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
