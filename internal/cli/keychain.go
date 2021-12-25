package cli

import (
	"colonies/pkg/security"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	keychainCmd.AddCommand(privateKeyCmd)
	rootCmd.AddCommand(keychainCmd)

	privateKeyCmd.Flags().StringVarP(&ID, "id", "", "", "Identity")
	privateKeyCmd.MarkFlagRequired("id")
}

var keychainCmd = &cobra.Command{
	Use:   "keychain",
	Short: "Manage private keys",
	Long:  "Manage private keys",
}

var privateKeyCmd = &cobra.Command{
	Use:   "privatekey",
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
