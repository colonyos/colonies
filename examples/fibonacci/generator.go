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
	coloniesHost := os.Getenv("COLONIES_SERVERHOST")
	coloniesPortStr := os.Getenv("COLONIES_SERVERPORT")
	coloniesPort, err := strconv.Atoi(coloniesPortStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	processSpec := core.CreateEmptyProcessSpec()
	processSpec.Conditions.ColonyID = colonyID
	processSpec.Conditions.RuntimeType = os.Getenv("COLONIES_RUNTIMETYPE")
	processSpec.Env["fibonacciNum"] = os.Args[1]

	client := client.CreateColoniesClient(coloniesHost, coloniesPort, true, false)
	addedProcess, err := client.SubmitProcessSpec(processSpec, runtimePrvKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Submitted a new process to the Colonies server with Id <" + addedProcess.ID + ">")
}
