package cli

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

func init() {
	colonyCmd.AddCommand(registerColonyCmd)
	colonyCmd.AddCommand(lsColoniesCmd)
	rootCmd.AddCommand(colonyCmd)

	colonyCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	colonyCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Server HTTP port")
	registerColonyCmd.Flags().StringVarP(&RootPassword, "rootpassword", "", "", "Root password to the Colonies server")
	registerColonyCmd.MarkFlagRequired("rootpassword")
	registerColonyCmd.Flags().StringVarP(&JSONSpecFile, "json", "", "", "JSON specification of a Colony")
	registerColonyCmd.MarkFlagRequired("json")

	lsColoniesCmd.Flags().StringVarP(&RootPassword, "rootpassword", "", "", "Root password to the Colonies server")
	lsColoniesCmd.MarkFlagRequired("rootpassword")
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
		jsonSpecBytes, err := ioutil.ReadFile(JSONSpecFile)
		CheckError(err)

		colony, err := core.ConvertJSONToColony(string(jsonSpecBytes))
		CheckError(err)

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		prvKey, err := security.GeneratePrivateKey()
		CheckError(err)

		colonyID, err := security.GenerateID(prvKey)
		CheckError(err)
		colony.SetID(colonyID)

		err = keychain.AddPrvKey(colonyID, prvKey)
		CheckError(err)

		addedColony, err := client.AddColony(colony, RootPassword, ServerHost, ServerPort)
		CheckError(err)

		fmt.Println(addedColony.ID())
	},
}

var lsColoniesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all Colonies",
	Long:  "List all Colonies",
	Run: func(cmd *cobra.Command, args []string) {
		coloniesFromServer, err := client.GetColonies(RootPassword, ServerHost, ServerPort)
		CheckError(err)

		jsonString, err := core.ConvertColonyArrayToJSON(coloniesFromServer)
		CheckError(err)

		fmt.Println(jsonString)
	},
}
