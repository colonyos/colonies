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
	colonyName := os.Getenv("COLONIES_COLONY_NAME")
	executorPrvKey := os.Getenv("COLONIES_EXECUTOR_PRVKEY")
	coloniesHost := os.Getenv("COLONIES_SERVER_HOST")
	coloniesPortStr := os.Getenv("COLONIES_SERVER_PORT")
	coloniesPort, err := strconv.Atoi(coloniesPortStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	// Ask the Colonies server to assign a process to this executor
	client := client.CreateColoniesClient(coloniesHost, coloniesPort, true, false)
	assignedProcess, err := client.Assign(colonyName, 100, "", "", executorPrvKey) // Max wait 100 seconds for assignment request
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

			attribute := core.CreateAttribute(assignedProcess.ID, colonyName, "", core.OUT, "result", fibonacci.String())
			client.AddAttribute(attribute, executorPrvKey)

			// Close the process as successful
			client.Close(assignedProcess.ID, executorPrvKey)
			return
		}
	}

	// Close the process as failed
	client.Fail(assignedProcess.ID, []string{"invalid arg"}, executorPrvKey)
}
