package cli

import (
	"errors"

	"github.com/colonyos/colonies/pkg/security"
	"github.com/colonyos/colonies/pkg/security/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	keychainCmd.AddCommand(addPrivateKeyCmd)
	keychainCmd.AddCommand(getPrivateKeyCmd)
	keychainCmd.AddCommand(genPrivateKeyCmd)
	rootCmd.AddCommand(keychainCmd)

	getPrivateKeyCmd.Flags().StringVarP(&ID, "id", "", "", "Identity")
	getPrivateKeyCmd.MarkFlagRequired("id")

	addPrivateKeyCmd.Flags().StringVarP(&ID, "id", "", "", "Identity")
	addPrivateKeyCmd.MarkFlagRequired("id")
	addPrivateKeyCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
	addPrivateKeyCmd.MarkFlagRequired("prvkey")
}

var keychainCmd = &cobra.Command{
	Use:   "keychain",
	Short: "Manage private keys",
	Long:  "Manage private keys",
}

var addPrivateKeyCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a private key",
	Long:  "Add a private key",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		err = keychain.AddPrvKey(ID, PrvKey)
		CheckError(err)
	},
}

var getPrivateKeyCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a private key for an Id",
	Long:  "Get a private key for an Id",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		privateKey, err := keychain.GetPrvKey(ID)
		if privateKey == "" {
			CheckError(errors.New("No private key found for identity <" + ID + ">"))
		}
		log.WithFields(log.Fields{"PrvKey": privateKey}).Info("Private key found in keychain")
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

		log.WithFields(log.Fields{"Id": id, "PrvKey": prvKey}).Info("Generated new private key and stored in keychain")
	},
}
