package cli

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	workflowCmd.AddCommand(submitWorkflowCmd)
	rootCmd.AddCommand(workflowCmd)

	submitWorkflowCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
	submitWorkflowCmd.Flags().StringVarP(&RuntimePrvKey, "runtimeprvkey", "", "", "Runtime private key")
	submitWorkflowCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony workflow")
	submitWorkflowCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	submitWorkflowCmd.Flags().BoolVarP(&Wait, "wait", "", false, "Colony Id")
	submitWorkflowCmd.MarkFlagRequired("spec")
}

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage workflows",
	Long:  "Manage workflows",
}

var submitWorkflowCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a workflow to a Colony",
	Long:  "Submit a workflow to a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		workflowSpec, err := core.ConvertJSONToWorkflowSpec(string(jsonSpecBytes))
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

		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)
		graph, err := client.SubmitWorkflowSpec(workflowSpec, RuntimePrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"ProcessGraphID": graph.ID}).Info("Workflow submitted")
	},
}
