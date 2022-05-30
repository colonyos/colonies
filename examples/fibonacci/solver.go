package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"

	fib "github.com/t-pwk/go-fibonacci"
)

func main() {
	colonyID := os.Getenv("COLONIES_COLONYID")
	runtimePrvKey := os.Getenv("COLONIES_RUNTIMEPRVKEY")
	coloniesHost := os.Getenv("COLONIES_SERVERHOST")
	coloniesPortStr := os.Getenv("COLONIES_SERVERPORT")
	coloniesPort, err := strconv.Atoi(coloniesPortStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// Ask the Colonies server to assign a process to this Runtime
	client := client.CreateColoniesClient(coloniesHost, coloniesPort, true, false)
	assignedProcess, err := client.AssignProcess(colonyID, runtimePrvKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Parse env attribute and calculate the given Fibonacci number
	for _, attribute := range assignedProcess.Attributes {
		if attribute.Key == "fibonacciNum" {
			fmt.Println("We were assigned process " + assignedProcess.ID)
			fmt.Println("Calculating Fibonacci serie for " + attribute.Value)
			nr, _ := strconv.Atoi(attribute.Value)
			fibonacci := fib.FibonacciBig(uint(nr))
			fmt.Println("Result: The last number in the Fibonacci serie " + attribute.Value + " is " + fibonacci.String())

			attribute := core.CreateAttribute(assignedProcess.ID, colonyID, core.OUT, "result", fibonacci.String())
			client.AddAttribute(attribute, runtimePrvKey)

			// Close the process as Successful
			client.CloseSuccessful(assignedProcess.ID, runtimePrvKey)
			return
		}
	}

	// Close the process as Failed
	client.CloseFailed(assignedProcess.ID, runtimePrvKey)
}
