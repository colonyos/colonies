package cli

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"colonies/pkg/security"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	attributeCmd.AddCommand(addAttributeCmd)
	attributeCmd.AddCommand(getAttributeCmd)
	rootCmd.AddCommand(attributeCmd)

	addAttributeCmd.Flags().StringVarP(&Key, "key", "", "", "Key")
	addAttributeCmd.MarkFlagRequired("key")
	addAttributeCmd.Flags().StringVarP(&Value, "value", "", "", "Value")
	addAttributeCmd.MarkFlagRequired("value")
	addAttributeCmd.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	addAttributeCmd.MarkFlagRequired("processid")
	addAttributeCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	addAttributeCmd.MarkFlagRequired("colonyid")
	addAttributeCmd.Flags().StringVarP(&ComputerID, "computerid", "", "", "Computer Id")
	addAttributeCmd.MarkFlagRequired("computerid")

	getAttributeCmd.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	getAttributeCmd.MarkFlagRequired("processid")
	getAttributeCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	getAttributeCmd.MarkFlagRequired("colonyid")
	getAttributeCmd.Flags().StringVarP(&ID, "id", "", "", "Colony or Computer Id")
	getAttributeCmd.MarkFlagRequired("id")
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
	Short: "Add an attribute to a proces",
	Long:  "Add an attribute to a process",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ComputerPrvKey == "" {
			ComputerPrvKey, err = keychain.GetPrvKey(ComputerID)
			CheckError(err)
		}

		attribute := core.CreateAttribute(ProcessID, core.OUT, Key, Value)
		addedAttribute, err := client.AddAttribute(attribute, ColonyID, ComputerPrvKey)
		CheckError(err)

		fmt.Println(addedAttribute.ToJSON())
	},
}

var getAttributeCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an attribute of a proces",
	Long:  "Get an attribute of a process",
	Run: func(cmd *cobra.Command, args []string) {
		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if PrvKey == "" {
			PrvKey, err = keychain.GetPrvKey(ID)
			CheckError(err)
		}

		attribute, err := client.GetAttribute(AttributeID, ProcessID, ColonyID, PrvKey)
		CheckError(err)

		fmt.Println(attribute.ToJSON())
	},
}
