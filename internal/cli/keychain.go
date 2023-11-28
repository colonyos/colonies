package cli

import (
	"github.com/colonyos/colonies/pkg/security/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	keychainCmd.AddCommand(genPrivateKeyCmd)
	rootCmd.AddCommand(keychainCmd)
}

var keychainCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage private keys",
	Long:  "Manage private keys",
}

var genPrivateKeyCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a private key",
	Long:  "Generate a private key",
	Run: func(cmd *cobra.Command, args []string) {
		crypto := crypto.CreateCrypto()
		prvKey, err := crypto.GeneratePrivateKey()
		CheckError(err)

		id, err := crypto.GenerateID(prvKey)
		CheckError(err)

		log.WithFields(log.Fields{"Id": id, "PrvKey": prvKey}).Info("Generated new private key")
	},
}
