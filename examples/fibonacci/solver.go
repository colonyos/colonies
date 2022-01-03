package main

import (
	"colonies/pkg/client"
	"fmt"
	"os"
	"strconv"

	fib "github.com/t-pwk/go-fibonacci"
)

func main() {
	colonyID := os.Getenv("COLONYID")
	runtimePrvKey := os.Args[1]

	// Ask the Colonies server to assign a process to this Runtime
	client := client.CreateColoniesClient("localhost", 8080, true)
	assignedProcess, err := client.AssignProcess(colonyID, runtimePrvKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Parse env attribute and calculate the given Fibonacci number
	for _, attribute := range assignedProcess.Attributes {
		if attribute.Key == "fibonacciNum" {
			nr, _ := strconv.Atoi(attribute.Value)
			fmt.Println(fib.FibonacciBig(uint(nr)))

			// Close the process as Successful
			client.MarkSuccessful(assignedProcess.ID, runtimePrvKey)
			return
		}
	}

	// Close the process as Failed
	client.MarkFailed(assignedProcess.ID, runtimePrvKey)
}
