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
	coloniesHost := os.Getenv("COLONIES_SERVER_HOST")
	coloniesPortStr := os.Getenv("COLONIES_SERVER_PORT")
	coloniesPort, err := strconv.Atoi(coloniesPortStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	client := client.CreateColoniesClient(coloniesHost, coloniesPort, true, false)

	// Subscribe for new processes
	subscription, err := client.SubscribeProcesses("cli", core.WAITING, 100, runtimePrvKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Waiting for processes to compute ...")
	go func() {
		for {
			select {
			case <-subscription.ProcessChan:
				assignedProcess, err := client.AssignProcess(colonyID, runtimePrvKey)
				if err != nil {
					fmt.Println(err)
					continue
				}

				// Parse env attribute and calculate the given Fibonacci number
				for _, attribute := range assignedProcess.Attributes {
					if attribute.Key == "fibonacciNum" {
						nr, _ := strconv.Atoi(attribute.Value)
						fmt.Println("Calculating Fibonacci serie for " + attribute.Value)
						fibonacci := fib.FibonacciBig(uint(nr))

						fmt.Println("We were assigned process " + assignedProcess.ID)
						fmt.Println("Result: The last number in the Fibonacci serie " + attribute.Value + " is " + fibonacci.String())

						attribute := core.CreateAttribute(assignedProcess.ID, colonyID, core.OUT, "result", fibonacci.String())
						client.AddAttribute(attribute, runtimePrvKey)

						// Close the process as Successful
						client.CloseSuccessful(assignedProcess.ID, runtimePrvKey)
						continue
					}
				}
			case err := <-subscription.ErrChan:
				fmt.Println(err)
			}
		}
	}()

	// Wait forever
	<-make(chan bool)
}
