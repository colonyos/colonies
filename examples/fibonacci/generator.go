package main

import (
	"colonies/pkg/client"
	"colonies/pkg/core"
	"fmt"
	"os"
)

func main() {
	colonyID := os.Getenv("COLONYID")
	runtimePrvKey := os.Args[1]
	fibonacciNum := os.Args[2]

	env := make(map[string]string)
	env["fibonacciNum"] = fibonacciNum

	processSpec := core.CreateProcessSpec(colonyID, []string{}, "FibonacciSolver", -1, 3, 1000, 10, 1, env)

	client := client.CreateColoniesClient("localhost", 8080, true)
	_, err := client.SubmitProcessSpec(processSpec, runtimePrvKey)
	if err != nil {
		fmt.Println(err)
	}
}
