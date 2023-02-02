package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
)

func main() {
	colonyID := os.Getenv("COLONIES_COLONY_ID")
	executorPrvKey := os.Getenv("COLONIES_EXECUTOR_PRVKEY")
	coloniesHost := os.Getenv("COLONIES_SERVER_HOST")
	coloniesPortStr := os.Getenv("COLONIES_SERVER_PORT")
	coloniesPort, err := strconv.Atoi(coloniesPortStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	processSpec := core.CreateEmptyProcessSpec()
	processSpec.Conditions.ColonyID = colonyID
	processSpec.Conditions.ExecutorType = os.Getenv("COLONIES_EXECUTOR_TYPE")
	processSpec.Env["fibonacciNum"] = os.Args[1]

	client := client.CreateColoniesClient(coloniesHost, coloniesPort, true, false)
	addedProcess, err := client.SubmitProcessSpec(processSpec, executorPrvKey)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Submitted a new process to the Colonies server with Id <" + addedProcess.ID + ">")
	fmt.Println("Waiting for process to be computed ...")

	fmt.Println(os.Getenv("COLONIES_EXECUTOR_TYPE"))

	subscription, _ := client.SubscribeProcess(addedProcess.ID, os.Getenv("COLONIES_EXECUTOR_TYPE"), core.SUCCESS, 100, executorPrvKey)
	process := <-subscription.ProcessChan

	for _, attribute := range process.Attributes {
		if attribute.Key == "result" {
			fmt.Println("Process was completed, the last number in the Fibonacci serie " + processSpec.Env["fibonacciNum"] + " is " + attribute.Value)
		}
	}
}
