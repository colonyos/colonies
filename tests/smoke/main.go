package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func submitProcess(client *client.ColoniesClient, colonyID string, executorPrvKey string) {
	funcSpec := core.FunctionSpec{
		Func:        "test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: 100,
		MaxExecTime: 2,
		MaxRetries:  10,
		Conditions:  core.Conditions{ColonyID: colonyID, ExecutorType: "bemisexecutor"},
		Env:         make(map[string]string)}

	client.Submit(&funcSpec, executorPrvKey)
}

func startCron(client *client.ColoniesClient, colonyID string, executorPrvKey string) {
	funcSpec1 := core.FunctionSpec{
		Name:        "cron_task1",
		Func:        "cron_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  10,
		Conditions:  core.Conditions{ColonyID: colonyID, ExecutorType: "bemisexecutor"},
		Env:         make(map[string]string)}

	funcSpec2 := core.FunctionSpec{
		Name:        "cron_task2",
		Func:        "cron_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  30,
		Conditions:  core.Conditions{ColonyID: colonyID, ExecutorType: "bemisexecutor"},
		Env:         make(map[string]string)}

	workflowSpec := core.CreateWorkflowSpec(colonyID)
	funcSpec2.AddDependency("cron_task1")
	workflowSpec.AddFunctionSpec(&funcSpec1)
	workflowSpec.AddFunctionSpec(&funcSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	checkError(err)

	cron := core.CreateCron(colonyID, "test_cron1"+core.GenerateRandomID(), "1 * * * * *", -1, false, jsonStr)
	_, err = client.AddCron(cron, executorPrvKey)
	checkError(err)
}

func startGenerator(client *client.ColoniesClient, colonyID string, executorPrvKey string) {
	funcSpec1 := core.FunctionSpec{
		Name:        "gen_task1",
		Func:        "gen_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  10,
		Conditions:  core.Conditions{ColonyID: colonyID, ExecutorType: "bemisexecutor"},
		Env:         make(map[string]string)}

	funcSpec2 := core.FunctionSpec{
		Name:        "gen_task2",
		Func:        "gen_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  30,
		Conditions:  core.Conditions{ColonyID: colonyID, ExecutorType: "bemisexecutor"},
		Env:         make(map[string]string)}

	workflowSpec := core.CreateWorkflowSpec(colonyID)
	funcSpec2.AddDependency("gen_task1")
	workflowSpec.AddFunctionSpec(&funcSpec1)
	workflowSpec.AddFunctionSpec(&funcSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	checkError(err)
	generator := core.CreateGenerator(colonyID, "test_genname"+core.GenerateRandomID(), jsonStr, 10)

	generator, err = client.AddGenerator(generator, executorPrvKey)
	checkError(err)

	go func() {
		for {
			client.PackGenerator(generator.ID, randStringRunes(10), executorPrvKey)
		}
	}()
}

func startExecutor(client *client.ColoniesClient, colonyID string, colonyPrvKey string) {
	crypto := crypto.CreateCrypto()
	executorPrvKey, err := crypto.GeneratePrivateKey()
	checkError(err)
	executorID, err := crypto.GenerateID(executorPrvKey)
	checkError(err)

	executor := core.CreateExecutor(executorID, "bemisexecutor", core.GenerateRandomID(), colonyID, time.Now(), time.Now())

	executor.Location.Long = 65.6120464058654 + rand.Float64()
	executor.Location.Lat = 22.132275667285477 + rand.Float64()

	_, err = client.AddExecutor(executor, colonyPrvKey)
	checkError(err)

	err = client.ApproveExecutor(executorID, colonyPrvKey)
	checkError(err)

	go func() {
		for {
			assignedProcess, err := client.Assign(colonyID, 10, executorPrvKey)
			if err == nil {
				time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
				client.Close(assignedProcess.ID, executorPrvKey)
			}
		}
	}()
}

func main() {
	colonyID := os.Getenv("COLONIES_COLONY_ID")
	colonyPrvKey := os.Getenv("COLONIES_COLONY_PRVKEY")

	serverHost := os.Getenv("COLONIES_SERVER_HOST")
	serverPortEnvStr := os.Getenv("COLONIES_SERVER_PORT")
	serverPort := -1
	var err error
	if serverPortEnvStr != "" {
		serverPort, err = strconv.Atoi(serverPortEnvStr)
		checkError(err)
	}

	tlsEnv := os.Getenv("COLONIES_TLS")
	insecure := true
	if tlsEnv == "true" {
		insecure = false
	} else if tlsEnv == "false" {
		insecure = true
	}

	executorPrvKey := os.Getenv("COLONIES_EXECUTOR_PRVKEY")

	client := client.CreateColoniesClient(serverHost, serverPort, insecure, true)

	for i := 0; i < 20; i++ {
		startExecutor(client, colonyID, colonyPrvKey)
	}

	startGenerator(client, colonyID, executorPrvKey)
	startCron(client, colonyID, executorPrvKey)

	//start := time.Now().UnixNano()
	for i := 0; i < 100000000; i++ {
		//now := time.Now().UnixNano()
		// delta := now - start
		// start = time.Now().UnixNano()
		//fmt.Println(i, " ", delta)

		submitProcess(client, colonyID, executorPrvKey)
	}

	done := make(chan struct{})
	<-done
}
