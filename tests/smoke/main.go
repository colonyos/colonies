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

func submitProcess(client *client.ColoniesClient, colonyID string, runtimePrvKey string) {
	processSpec := core.ProcessSpec{
		Func:        "test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: 100,
		MaxExecTime: 2,
		MaxRetries:  10,
		Conditions:  core.Conditions{ColonyID: colonyID, RuntimeType: "bemisworker"},
		Env:         make(map[string]string)}

	client.SubmitProcessSpec(&processSpec, runtimePrvKey)
}

func startCron(client *client.ColoniesClient, colonyID string, runtimePrvKey string) {
	processSpec1 := core.ProcessSpec{
		Name:        "cron_task1",
		Func:        "cron_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  10,
		Conditions:  core.Conditions{ColonyID: colonyID, RuntimeType: "bemisworker"},
		Env:         make(map[string]string)}

	processSpec2 := core.ProcessSpec{
		Name:        "cron_task2",
		Func:        "cron_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  30,
		Conditions:  core.Conditions{ColonyID: colonyID, RuntimeType: "bemisworker"},
		Env:         make(map[string]string)}

	workflowSpec := core.CreateWorkflowSpec(colonyID)
	processSpec2.AddDependency("cron_task1")
	workflowSpec.AddProcessSpec(&processSpec1)
	workflowSpec.AddProcessSpec(&processSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	checkError(err)

	cron := core.CreateCron(colonyID, "test_cron1"+core.GenerateRandomID(), "1 * * * * *", -1, false, jsonStr)
	_, err = client.AddCron(cron, runtimePrvKey)
	checkError(err)
}

func startGenerator(client *client.ColoniesClient, colonyID string, runtimePrvKey string) {
	processSpec1 := core.ProcessSpec{
		Name:        "gen_task1",
		Func:        "gen_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  10,
		Conditions:  core.Conditions{ColonyID: colonyID, RuntimeType: "bemisworker"},
		Env:         make(map[string]string)}

	processSpec2 := core.ProcessSpec{
		Name:        "gen_task2",
		Func:        "gen_test_func",
		Args:        []string{"arg1"},
		MaxWaitTime: -1,
		MaxExecTime: 2,
		MaxRetries:  30,
		Conditions:  core.Conditions{ColonyID: colonyID, RuntimeType: "bemisworker"},
		Env:         make(map[string]string)}

	workflowSpec := core.CreateWorkflowSpec(colonyID)
	processSpec2.AddDependency("gen_task1")
	workflowSpec.AddProcessSpec(&processSpec1)
	workflowSpec.AddProcessSpec(&processSpec2)
	jsonStr, err := workflowSpec.ToJSON()
	checkError(err)
	generator := core.CreateGenerator(colonyID, "test_genname"+core.GenerateRandomID(), jsonStr, 10)

	generator, err = client.AddGenerator(generator, runtimePrvKey)
	checkError(err)

	go func() {
		for {
			client.PackGenerator(generator.ID, randStringRunes(10), runtimePrvKey)
		}
	}()
}

func startWorker(client *client.ColoniesClient, colonyID string, colonyPrvKey string) {
	crypto := crypto.CreateCrypto()
	runtimePrvKey, err := crypto.GeneratePrivateKey()
	checkError(err)
	runtimeID, err := crypto.GenerateID(runtimePrvKey)
	checkError(err)

	runtime := core.CreateRuntime(runtimeID, "bemisworker", core.GenerateRandomID(), colonyID, "AMD Ryzen 9 5950X (32) @ 3.400GHz", 32, 80326, "NVIDIA GeForce RTX 2080 Ti Rev. A", 1, time.Now(), time.Now())

	runtime.Location.Long = 65.6120464058654 + rand.Float64()
	runtime.Location.Lat = 22.132275667285477 + rand.Float64()

	_, err = client.AddRuntime(runtime, colonyPrvKey)
	checkError(err)

	err = client.ApproveRuntime(runtimeID, colonyPrvKey)
	checkError(err)

	go func() {
		for {
			assignedProcess, err := client.AssignProcess(colonyID, 10, runtimePrvKey)
			if err == nil {
				time.Sleep(time.Duration(rand.Intn(300)) * time.Millisecond)
				client.Close(assignedProcess.ID, runtimePrvKey)
			}
		}
	}()
}

func main() {
	colonyID := os.Getenv("COLONIES_COLONYID")
	colonyPrvKey := os.Getenv("COLONIES_COLONYPRVKEY")

	serverHost := os.Getenv("COLONIES_SERVERHOST")
	serverPortEnvStr := os.Getenv("COLONIES_SERVERPORT")
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

	runtimePrvKey := os.Getenv("COLONIES_RUNTIMEPRVKEY")

	client := client.CreateColoniesClient(serverHost, serverPort, insecure, true)

	for i := 0; i < 20; i++ {
		startWorker(client, colonyID, colonyPrvKey)
	}

	startGenerator(client, colonyID, runtimePrvKey)
	startCron(client, colonyID, runtimePrvKey)

	//start := time.Now().UnixNano()
	for i := 0; i < 100000000; i++ {
		//now := time.Now().UnixNano()
		// delta := now - start
		// start = time.Now().UnixNano()
		//fmt.Println(i, " ", delta)

		submitProcess(client, colonyID, runtimePrvKey)
	}

	done := make(chan struct{})
	<-done
}
