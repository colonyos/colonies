package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security"
	"github.com/kataras/tablewriter"
	"github.com/spf13/cobra"
)

func init() {
	attributeCmd.AddCommand(addAttributeCmd)
	attributeCmd.AddCommand(getAttributeCmd)
	rootCmd.AddCommand(attributeCmd)

	attributeCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	attributeCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Server HTTP port")

	addAttributeCmd.Flags().StringVarP(&Key, "key", "", "", "Key")
	addAttributeCmd.MarkFlagRequired("key")
	addAttributeCmd.Flags().StringVarP(&Value, "value", "", "", "Value")
	addAttributeCmd.MarkFlagRequired("value")
	addAttributeCmd.Flags().StringVarP(&ProcessID, "processid", "", "", "Process Id")
	addAttributeCmd.MarkFlagRequired("processid")
	addAttributeCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	addAttributeCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")

	getAttributeCmd.Flags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	getAttributeCmd.Flags().StringVarP(&RuntimeID, "runtimeid", "", "", "Runtime Id")
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
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		attribute := core.CreateAttribute(ProcessID, ColonyID, core.OUT, Key, Value)
		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		addedAttribute, err := client.AddAttribute(attribute, RuntimePrvKey)
		CheckError(err)

		fmt.Println(addedAttribute.ID)
	},
}

var getAttributeCmd = &cobra.Command{
	Use:   "get",
	Short: "Get an attribute of a proces",
	Long:  "Get an attribute of a process",
	Run: func(cmd *cobra.Command, args []string) {
		parseServerEnv()

		keychain, err := security.CreateKeychain(KEYCHAIN_PATH)
		CheckError(err)

		if ColonyID == "" {
			ColonyID = os.Getenv("COLONYID")
		}
		if ColonyID == "" {
			CheckError(errors.New("Unknown Colony Id"))
		}

		if RuntimeID == "" {
			RuntimeID = os.Getenv("RUNTIMEID")
		}
		if RuntimeID == "" {
			CheckError(errors.New("Unknown Runtime Id"))
		}

		if RuntimePrvKey == "" {
			RuntimePrvKey, err = keychain.GetPrvKey(RuntimeID)
			CheckError(err)
		}

		client := client.CreateColoniesClient(ServerHost, ServerPort, true) // XXX: Insecure
		attribute, err := client.GetAttribute(AttributeID, RuntimePrvKey)
		CheckError(err)

		var attributeType string
		switch attribute.AttributeType {
		case core.IN:
			attributeType = "In"
		case core.OUT:
			attributeType = "Out"
		case core.ERR:
			attributeType = "Err"
		case core.ENV:
			attributeType = "Env"
		default:
			attributeType = "Unknown"
		}

		var key string
		if len(attribute.Key) > MaxAttributeLength {
			key = attribute.Key[0:MaxAttributeLength] + "..."
		} else {
			key = attribute.Key
		}

		var value string
		if len(attribute.Value) > MaxAttributeLength {
			value = attribute.Value[0:MaxAttributeLength] + "..."
		} else {
			value = attribute.Value
		}

		attributeData := [][]string{
			[]string{"ID", attribute.ID},
			[]string{"TargetID", attribute.TargetID},
			[]string{"AttributeType", attributeType},
			[]string{"Key", key},
			[]string{"Value", value},
		}
		attributeTable := tablewriter.NewWriter(os.Stdout)
		for _, v := range attributeData {
			attributeTable.Append(v)
		}
		attributeTable.SetAlignment(tablewriter.ALIGN_LEFT)
		attributeTable.Render()
	},
}
