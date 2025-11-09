package cli

import (
	"fmt"
	"os"

	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	nodeCmd.AddCommand(lsNodesCmd)
	nodeCmd.AddCommand(getNodeCmd)
	rootCmd.AddCommand(nodeCmd)

	nodeCmd.PersistentFlags().StringVarP(&ColonyName, "colonyid", "", "", "Colony Id")
	nodeCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	nodeCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	lsNodesCmd.Flags().BoolVarP(&JSON, "json", "", false, "Print JSON instead of tables")
	lsNodesCmd.Flags().StringVarP(&Location, "location", "", "", "Filter by location")

	getNodeCmd.Flags().StringVarP(&NodeName, "name", "", "", "Node name")
	getNodeCmd.MarkFlagRequired("name")
}

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage nodes",
	Long:  "Manage nodes",
}

var lsNodesCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all nodes",
	Long:  "List all nodes",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		var nodes []*core.Node
		var err error

		if Location != "" {
			nodes, err = client.GetNodesByLocation(ColonyName, Location, PrvKey)
			CheckError(err)
		} else {
			nodes, err = client.GetNodes(ColonyName, PrvKey)
			CheckError(err)
		}

		if len(nodes) == 0 {
			log.Info("No nodes found")
			os.Exit(0)
		}

		if JSON {
			jsonString, err := core.ConvertNodesToJSON(nodes)
			CheckError(err)
			fmt.Println(jsonString)
			os.Exit(0)
		}

		// Get all executors to count them per node
		executors, err := client.GetExecutors(ColonyName, PrvKey)
		CheckError(err)

		printNodesTable(nodes, executors)
	},
}

var getNodeCmd = &cobra.Command{
	Use:   "get",
	Short: "Get node details",
	Long:  "Get node details",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		node, err := client.GetNode(ColonyName, NodeName, PrvKey)
		CheckError(err)

		if node == nil {
			log.Error("Node not found")
			os.Exit(-1)
		}

		if JSON {
			jsonString, err := node.ToJSON()
			CheckError(err)
			fmt.Println(jsonString)
			os.Exit(0)
		}

		// Get all executors to show which ones are on this node
		executors, err := client.GetExecutors(ColonyName, PrvKey)
		CheckError(err)

		printNodeTable(node, executors)
	},
}
