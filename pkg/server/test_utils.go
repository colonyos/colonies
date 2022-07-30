package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/etcd"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/stretchr/testify/assert"
)

type testEnv1 struct {
	colony1PrvKey  string
	colony1ID      string
	colony2PrvKey  string
	colony2ID      string
	runtime1PrvKey string
	runtime1ID     string
	runtime2PrvKey string
	runtime2ID     string
}

type testEnv2 struct {
	colonyID      string
	colony        *core.Colony
	colonyPrvKey  string
	runtimeID     string
	runtime       *core.Runtime
	runtimePrvKey string
}

const EnableTLS = true
const Insecure = false
const SkipTLSVerify = true

func setupTestEnv1(t *testing.T) (*testEnv1, *client.ColoniesClient, *ColoniesServer, string, chan bool) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony1, colony1PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	runtime1, runtime1PrvKey, err := utils.CreateTestRuntimeWithKey(colony1.ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime1, colony1PrvKey)
	assert.Nil(t, err)

	runtime2, runtime2PrvKey, err := utils.CreateTestRuntimeWithKey(colony2.ID)
	assert.Nil(t, err)
	_, err = client.AddRuntime(runtime2, colony2PrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime1.ID, colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime2.ID, colony2PrvKey)
	assert.Nil(t, err)

	env := &testEnv1{colony1PrvKey: colony1PrvKey,
		colony1ID:      colony1.ID,
		colony2PrvKey:  colony2PrvKey,
		colony2ID:      colony2.ID,
		runtime1PrvKey: runtime1PrvKey,
		runtime1ID:     runtime1.ID,
		runtime2PrvKey: runtime2PrvKey,
		runtime2ID:     runtime2.ID}

	return env, client, server, serverPrvKey, done
}

func setupTestEnv2(t *testing.T) (*testEnv2, *client.ColoniesClient, *ColoniesServer, string, chan bool) {
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	runtime, runtimePrvKey, err := utils.CreateTestRuntimeWithKey(colony.ID)
	_, err = client.AddRuntime(runtime, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveRuntime(runtime.ID, colonyPrvKey)
	assert.Nil(t, err)

	env := &testEnv2{colonyID: colony.ID,
		colony:        colony,
		colonyPrvKey:  colonyPrvKey,
		runtimeID:     runtime.ID,
		runtime:       runtime,
		runtimePrvKey: runtimePrvKey}

	return env, client, server, serverPrvKey, done
}

func prepareTests(t *testing.T) (*client.ColoniesClient, *ColoniesServer, string, chan bool) {
	client := client.CreateColoniesClient(TESTHOST, TESTPORT, Insecure, SkipTLSVerify)

	debug := false
	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	crypto := crypto.CreateCrypto()
	serverPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := crypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)

	node := etcd.Node{Name: "etcd", Host: "localhost", ClientPort: 24100, PeerPort: 23100}
	cluster := etcd.Cluster{}
	cluster.AddNode(node)
	server := CreateColoniesServer(db, TESTPORT, serverID, EnableTLS, "../../cert/key.pem", "../../cert/cert.pem", debug, node, cluster, "/tmp/colonies/etcd")

	done := make(chan bool)
	go func() {
		server.ServeForever()
		db.Close()
		done <- true
	}()

	return client, server, serverPrvKey, done
}

func generateDiamondtWorkflowSpec(colonyID string) *core.WorkflowSpec {
	//         task1
	//          / \
	//     task2   task3
	//          \ /
	//         task4

	workflowSpec := core.CreateWorkflowSpec(colonyID)

	processSpec1 := core.CreateEmptyProcessSpec()
	processSpec1.Name = "task1"
	processSpec1.Conditions.ColonyID = colonyID
	processSpec1.Conditions.RuntimeType = "test_runtime_type"

	processSpec2 := core.CreateEmptyProcessSpec()
	processSpec2.Name = "task2"
	processSpec2.Conditions.ColonyID = colonyID
	processSpec2.Conditions.RuntimeType = "test_runtime_type"

	processSpec3 := core.CreateEmptyProcessSpec()
	processSpec3.Name = "task3"
	processSpec3.Conditions.ColonyID = colonyID
	processSpec3.Conditions.RuntimeType = "test_runtime_type"

	processSpec4 := core.CreateEmptyProcessSpec()
	processSpec4.Name = "task4"
	processSpec4.Conditions.ColonyID = colonyID
	processSpec4.Conditions.RuntimeType = "test_runtime_type"

	processSpec2.AddDependency("task1")
	processSpec3.AddDependency("task1")
	processSpec4.AddDependency("task2")
	processSpec4.AddDependency("task3")

	workflowSpec.AddProcessSpec(processSpec1)
	workflowSpec.AddProcessSpec(processSpec2)
	workflowSpec.AddProcessSpec(processSpec3)
	workflowSpec.AddProcessSpec(processSpec4)

	return workflowSpec
}

func waitForProcesses(t *testing.T, server *ColoniesServer, processes []*core.Process, state int) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancelCtx()
	wait := make(chan error)
	for _, process := range processes {
		go func(process *core.Process) {
			_, err := server.controller.eventHandler.waitForProcess(process.ProcessSpec.Conditions.RuntimeType, state, process.ID, ctx)
			fmt.Println(process.ID, " state:", process.State)
			wait <- err
		}(process)
	}

	var err error
	for i := 0; i < len(processes); i++ {
		err = <-wait
		assert.Nil(t, err)
	}
}
