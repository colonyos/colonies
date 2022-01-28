package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/kataras/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	colonyCmd.AddCommand(registerColonyCmd)
	colonyCmd.AddCommand(unregisterColonyCmd)
	colonyCmd.AddCommand(lsColoniesCmd)
	rootCmd.AddCommand(colonyCmd)

	colonyCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	colonyCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Server HTTP port")

	registerColonyCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	registerColonyCmd.MarkFlagRequired("serverid")
	registerColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	registerColonyCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony")
	registerColonyCmd.MarkFlagRequired("spec")

	unregisterColonyCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	unregisterColonyCmd.MarkFlagRequired("serverid")
	unregisterColonyCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
	unregisterColonyCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	unregisterColonyCmd.MarkFlagRequired("colonyid")

	lsColoniesCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	lsColoniesCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
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

		if ServerPrvKey == "" {
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		addedColony, err := client.AddColony(colony, ServerPrvKey)
		CheckError(err)

		err = keychain.AddPrvKey(colonyID, prvKey)
		CheckError(err)

		fmt.Println(addedColony.ID)
	},
}

var unregisterColonyCmd = &cobra.Command{
	Use:   "unregister",
	Short: "Unregister a Colony",
	Long:  "Unregister a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ServerPrvKey == "" {
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		err = client.DeleteColony(ColonyID, ServerPrvKey)
		CheckError(err)

		fmt.Println("Colony with Id <" + ColonyID + "> was successfully unregistered")
	},
}

var lsColoniesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all Colonies",
	Long:  "List all Colonies",
	Run: func(cmd *cobra.Command, args []string) {
		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

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
