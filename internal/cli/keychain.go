package cli

import (
	"fmt"
	"os"

	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/spf13/cobra"
)

func init() {
	keychainCmd.AddCommand(getPrivateKeyCmd)
	keychainCmd.AddCommand(genPrivateKeyCmd)
	rootCmd.AddCommand(keychainCmd)

	getPrivateKeyCmd.Flags().StringVarP(&ID, "id", "", "", "Identity")
	getPrivateKeyCmd.MarkFlagRequired("id")
}

var keychainCmd = &cobra.Command{
	Use:   "keychain",
	Short: "Manage private keys",
	Long:  "Manage private keys",
}

var getPrivateKeyCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a private key for an identity",
	Long:  "Get a private key for an identity",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		privateKey, err := keychain.GetPrvKey(ID)
		if privateKey == "" {
			fmt.Println("No private key found for identity <" + ID + ">")
			os.Exit(-1)
		}
		fmt.Println(privateKey)
	},
}

var genPrivateKeyCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a private key",
	Long:  "Generate a private key",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		crypto := crypto.CreateCrypto()
		prvKey, err := crypto.GeneratePrivateKey()
		CheckError(err)

		id, err := crypto.GenerateID(prvKey)
		CheckError(err)

		err = keychain.AddPrvKey(id, prvKey)
		CheckError(err)

		fmt.Println("Id: " + id)
		fmt.Println("PrvKey: " + prvKey)
	},
}
