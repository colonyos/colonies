package cli

import (
	"errors"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/kataras/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	clusterCmd.AddCommand(clusterInfoCmd)
	rootCmd.AddCommand(clusterCmd)

	clusterInfoCmd.Flags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	clusterInfoCmd.Flags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")
	clusterInfoCmd.Flags().StringVarP(&ServerID, "serverid", "", "", "Colonies server Id")
	clusterInfoCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage Colonies clusters",
	Long:  "Manage Colonies clusters",
}

var clusterInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show info about a colonies cluster",
	Long:  "Show info about a colonies cluster",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ServerID == "" {
			ServerID = os.Getenv("COLONIES_SERVERID")
		}
		if ServerID == "" {
			CheckError(errors.New("Unknown Server Id"))
		}

		if ServerPrvKey == "" {
			ServerPrvKey, err = keychain.GetPrvKey(ServerID)
			CheckError(err)
		}

		log.WithFields(log.Fields{"ServerHost": ServerHost, "ServerPort": ServerPort, "Insecure": Insecure}).Info("Starting a Colonies client")
		client := client.CreateColoniesClient(ServerHost, ServerPort, Insecure, SkipTLSVerify)

		cluster, err := client.GetClusterInfo(ServerPrvKey)
		CheckError(err)

		var data [][]string
		for _, node := range cluster.Nodes {
			data = append(data, []string{node.Name,
				node.Host,
				strconv.Itoa(node.APIPort),
				strconv.Itoa(node.EtcdClientPort),
				strconv.Itoa(node.EtcdPeerPort),
				strconv.Itoa(node.RelayPort),
				isLeader(cluster.Leader.Name, node.Name)})
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Host", "APIPort", "EtcdClientPort", "EtcdPeerPort", "RelayPort", "Leader"})
		for _, v := range data {
			table.Append(v)
		}
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.Render()
	},
}

func isLeader(leader string, name string) string {
	if leader == name {
		return "True"
	}
	return "False"
}
