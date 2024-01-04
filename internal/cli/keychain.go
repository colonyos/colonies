package cli

import (
	icrypto "github.com/colonyos/colonies/internal/crypto"
	"github.com/colonyos/colonies/pkg/security/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	keychainCmd.AddCommand(genPrivateKeyCmd)
	keychainCmd.AddCommand(idCmd)
	rootCmd.AddCommand(keychainCmd)

	idCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
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

var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Show the Id for a given private key",
	Long:  "Show the Id for a given private key",
	Run: func(cmd *cobra.Command, args []string) {
		id, err := icrypto.GenerateID(PrvKey)
		CheckError(err)
		log.WithFields(log.Fields{"Id": id}).Info("Corresponding Id for the given private key")
	},
}
