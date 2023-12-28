package cli

import (
	"github.com/spf13/cobra"
)

func init() {
	clusterCmd.AddCommand(clusterInfoCmd)
	rootCmd.AddCommand(clusterCmd)

	clusterInfoCmd.Flags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	clusterInfoCmd.Flags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")
	clusterInfoCmd.Flags().StringVarP(&ServerPrvKey, "serverprvkey", "", "", "Colonies server private key")
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Manage clusters",
	Long:  "Manage clusters",
}

var clusterInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show info about a Colonies cluster",
	Long:  "Show info about a Colonies cluster",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		cluster, err := client.GetClusterInfo(ServerPrvKey)
		CheckError(err)

		printClusterInfoTable(cluster)
	},
}

func isLeader(leader string, name string) string {
	if leader == name {
		return "True"
	}
	return "False"
}
