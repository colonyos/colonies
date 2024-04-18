package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
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
	executorType := os.Getenv("COLONIES_EXECUTOR_TYPE")

	funcSpec := core.CreateEmptyFunctionSpec()
	funcSpec.Conditions.ColonyName = colonyName
	funcSpec.Conditions.ExecutorType = executorType
	funcSpec.Env["fibonacciNum"] = os.Args[1]

	client := client.CreateColoniesClient(coloniesHost, coloniesPort, true, false)
	addedProcess, err := client.Submit(funcSpec, executorPrvKey)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Submitted a new process to the Colonies server with Id <" + addedProcess.ID + ">")
	fmt.Println("Waiting for process to be computed ...")

	fmt.Println(executorType)

	subscription, _ := client.SubscribeProcess(colonyName, addedProcess.ID, executorType, core.SUCCESS, 100, executorPrvKey)
	process := <-subscription.ProcessChan

	for _, attribute := range process.Attributes {
		if attribute.Key == "result" {
			fmt.Println("Process was completed, the last number in the Fibonacci serie " + funcSpec.Env["fibonacciNum"] + " is " + attribute.Value)
		}
	}
}
