package cli

import (
	"github.com/colonyos/colonies/pkg/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	attributeCmd.AddCommand(addAttributeCmd)
	attributeCmd.AddCommand(getAttributeCmd)
	rootCmd.AddCommand(attributeCmd)

	attributeCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	attributeCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", -1, "Server HTTP port")

	addAttributeCmd.Flags().StringVarP(&PrvKey, "prvkey", "", "", "Executor private key")
	addAttributeCmd.Flags().StringVarP(&Key, "key", "", "", "Key")
	addAttributeCmd.MarkFlagRequired("key")
	addAttributeCmd.Flags().StringVarP(&Value, "value", "", "", "Value")
	addAttributeCmd.MarkFlagRequired("value")
	addAttributeCmd.Flags().StringVarP(&ProcessID, "processid", "p", "", "Process Id")
	addAttributeCmd.MarkFlagRequired("processid")

	getAttributeCmd.Flags().StringVarP(&AttributeID, "attributeid", "", "", "Attribute Id")
	getAttributeCmd.MarkFlagRequired("attributeid")
}

var attributeCmd = &cobra.Command{
	Use:   "attribute",
	Short: "Manage process attributes",
	Long:  "Manage process attributes",
}

var addAttributeCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an attribute to a process",
	Long:  "Add an attribute to a process",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		process, err := client.GetProcess(ProcessID, PrvKey)
		CheckError(err)

		attribute := core.CreateAttribute(ProcessID, ColonyName, process.ProcessGraphID, core.OUT, Key, Value)

		addedAttribute, err := client.AddAttribute(attribute, PrvKey)
		CheckError(err)

		log.WithFields(log.Fields{"AttributeID": addedAttribute.ID}).Info("Attribute added")
	},
}

var getAttributeCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an attribute of a process",
	Long:  "Get an attribute of a process",
	Run: func(cmd *cobra.Command, args []string) {
		client := setup()

		attribute, err := client.GetAttribute(AttributeID, PrvKey)
		CheckError(err)

		printAttributeTable(&attribute)
	},
}
