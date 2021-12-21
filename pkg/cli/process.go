package cli

import (
	"colonies/pkg/core"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

func init() {
	processCmd.AddCommand(submitProcessCmd)
	rootCmd.AddCommand(processCmd)

	processCmd.PersistentFlags().StringVarP(&ColonyID, "colonyid", "", "", "Colony Id")
	processCmd.PersistentFlags().StringVarP(&ServerHost, "host", "", "localhost", "Server host")
	processCmd.PersistentFlags().IntVarP(&ServerPort, "port", "", 8080, "Server HTTP port")

	submitProcessCmd.Flags().StringVarP(&SpecFile, "spec", "", "", "JSON specification of a Colony process")
	approveComputersCmd.MarkFlagRequired("spec")
	submitProcessCmd.Flags().StringVarP(&InputFile, "in", "", "", "JSON specification of the input data")
	approveComputersCmd.MarkFlagRequired("in")
}

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Manage Colony processes",
	Long:  "Manage Colony processes",
}

var submitProcessCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a process to a Colony",
	Long:  "Submit a process to a Colony",
	Run: func(cmd *cobra.Command, args []string) {
		jsonSpecBytes, err := ioutil.ReadFile(SpecFile)
		CheckError(err)

		inputBytes, err := ioutil.ReadFile(InputFile)
		CheckError(err)

		process, err := core.ConvertJSONToProcess(string(jsonSpecBytes))
		CheckError(err)

		fmt.Println(string(inputBytes))
		fmt.Println(process)
		//fmt.Println(process.ToJSON())
	},
}
