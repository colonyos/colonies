package cli

import (
	"encoding/hex"

	icrypto "github.com/colonyos/colonies/internal/crypto"
	"github.com/colonyos/colonies/pkg/security/crypto"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	securityCmd.AddCommand(genPrivateKeyCmd)
	securityCmd.AddCommand(idCmd)
	securityCmd.AddCommand(genLibP2PIDCmd)
	rootCmd.AddCommand(securityCmd)

	idCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Private key")
}

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Manage security and cryptographic keys",
	Long:  "Manage security and cryptographic keys",
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

var genLibP2PIDCmd = &cobra.Command{
	Use:   "generatep2pid",
	Short: "Generate a LibP2P identity (private key)",
	Long:  "Generate a LibP2P identity that can be used with COLONIES_LIBP2P_IDENTITY environment variable",
	Run: func(cmd *cobra.Command, args []string) {
		// Generate crypto from libp2p
		privKey, _, err := libp2pcrypto.GenerateEd25519Key(nil)
		CheckError(err)

		// Marshal the private key to bytes
		privKeyBytes, err := libp2pcrypto.MarshalPrivateKey(privKey)
		CheckError(err)

		// Get the peer ID from the private key
		peerID, err := peer.IDFromPrivateKey(privKey)
		CheckError(err)

		// Log the results
		log.WithFields(log.Fields{
			"PeerID":  peerID.String(),
			"PrvKey":  hex.EncodeToString(privKeyBytes),
			"Example": "/ip4/127.0.0.1/tcp/5000/p2p/" + peerID.String(),
		}).Info("Generated new LibP2P identity")
	},
}
