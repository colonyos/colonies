package cli

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"colonies/pkg/security/crypto"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kataras/tablewriter"
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
	registerColonyCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony")
	registerColonyCmd.MarkFlagRequired("spec")

	lsColoniesCmd.Flags().StringVarP(&RootPassword, "rootpassword", "", "", "Root password to the Colonies server")
	lsColoniesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
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
		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		colony, err := core.ConvertJSONToColony(string(jsonSpecBytes))
		CheckError(err)

		crypto := crypto.CreateCrypto()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		prvKey, err := crypto.GeneratePrivateKey()
		CheckError(err)

		colonyID, err := crypto.GenerateID(prvKey)
		CheckError(err)
		colony.SetID(colonyID)

		err = keychain.AddPrvKey(colonyID, prvKey)
		CheckError(err)

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		addedColony, err := client.AddColony(colony, RootPassword)
		CheckError(err)

		fmt.Println(addedColony.ID)
	},
}

var lsColoniesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all Colonies",
	Long:  "List all Colonies",
	Run: func(cmd *cobra.Command, args []string) {
		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		coloniesFromServer, err := client.GetColonies(RootPassword)
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
