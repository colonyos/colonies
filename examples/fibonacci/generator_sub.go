package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
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

	env := make(map[string]string)
	env["fibonacciNum"] = os.Args[1]

	processSpec := core.CreateProcessSpec("", "", []string{}, []string{}, []string{}, colonyID, []string{}, "cli", -1, 3, 1000, 10, 1, env)

	client := client.CreateColoniesClient(coloniesHost, coloniesPort, true, false)
	addedProcess, err := client.SubmitProcessSpec(processSpec, runtimePrvKey)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Submitted a new process to the Colonies server with Id <" + addedProcess.ID + ">")
	fmt.Println("Waiting for process to be computed ...")

	subscription, _ := client.SubscribeProcess(addedProcess.ID, core.SUCCESS, 100, runtimePrvKey)
	process := <-subscription.ProcessChan

	for _, attribute := range process.Attributes {
		if attribute.Key == "result" {
			fmt.Println("Process was completed, the last number in the Fibonacci serie " + env["fibonacciNum"] + "is " + attribute.Value)
		}
	}
}
