package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/database/postgresql"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
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
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
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
	os.RemoveAll("/tmp/colonies")
	client := client.CreateColoniesClient(TESTHOST, TESTPORT, Insecure, SkipTLSVerify)

	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	crypto := crypto.CreateCrypto()
	serverPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := crypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)

	node := cluster.Node{Name: "etcd", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: TESTPORT}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	server := CreateColoniesServer(db, TESTPORT, serverID, EnableTLS, "../../cert/key.pem", "../../cert/cert.pem", node, clusterConfig, "/tmp/colonies/etcd")

	done := make(chan bool)
	go func() {
		server.ServeForever()
		db.Close()
		done <- true
	}()

	return client, server, serverPrvKey, done
}

func createTestColoniesController(db database.Database) *coloniesController {
	node := cluster.Node{Name: "etcd", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: TESTPORT}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	return createColoniesController(db, node, clusterConfig, "/tmp/colonies/etcd")
}

func createTestColoniesController2(db database.Database) *coloniesController {
	node := cluster.Node{Name: "etcd2", Host: "localhost", EtcdClientPort: 26100, EtcdPeerPort: 27100, RelayPort: 28100, APIPort: TESTPORT}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	return createColoniesController(db, node, clusterConfig, "/tmp/colonies/etcd")
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
			wait <- err
		}(process)
	}

	var err error
	for i := 0; i < len(processes); i++ {
		err = <-wait
		assert.Nil(t, err)
	}
}

func verifyRPCReplyMsgHasErr(t *testing.T, b []byte) {
	rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(string(b))
	assert.Nil(t, err)
	assert.True(t, rpcReplyMsg.Error)
}

// Cluster testing

type ServerInfo struct {
	ServerID     string
	ServerPrvKey string
	Server       *ColoniesServer
	Node         cluster.Node
	Done         chan struct{}
}

func StartCluster(t *testing.T, db database.Database, size int) []ServerInfo {
	os.RemoveAll("/tmp/colonies")
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	clusterConfig := cluster.Config{}
	for i := 0; i < size; i++ {
		node := cluster.Node{
			Name:           "etcd" + strconv.Itoa(i),
			Host:           "localhost",
			EtcdClientPort: 21000 + i,
			EtcdPeerPort:   22000 + i,
			RelayPort:      23000 + i,
			APIPort:        24000 + i}
		clusterConfig.AddNode(node)
	}

	crypto := crypto.CreateCrypto()
	serverPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := crypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)

	sChan := make(chan ServerInfo)
	for i, node := range clusterConfig.Nodes {
		go func(i int, node cluster.Node) {
			log.WithFields(log.Fields{"APIPort": node.APIPort}).Info("Starting ColoniesServer")
			server := CreateColoniesServer(db, node.APIPort, serverID, false, "", "", node, clusterConfig, "/tmp/colonies/etcd"+strconv.Itoa(i))
			done := make(chan struct{})
			s := ServerInfo{ServerID: serverID, ServerPrvKey: serverPrvKey, Server: server, Node: node, Done: done}
			go func(i int) {
				log.Info("ColoniesServer serving")
				server.ServeForever()
				log.Info("ColoniesServer stopped")
				done <- struct{}{}
			}(i)
			sChan <- s
		}(i, node)
	}

	var servers []ServerInfo
	for range clusterConfig.Nodes {
		s := <-sChan
		servers = append(servers, s)
	}

	return servers
}

func WaitForCluster(t *testing.T, cluster []ServerInfo) {
	serverReady := 0
	for {
		for _, s := range cluster {
			client := client.CreateColoniesClient("localhost", s.Node.APIPort, true, true)
			err := client.CheckHealth()
			if err == nil {
				serverReady++
			} else {
				time.Sleep(50 * time.Millisecond)
				fmt.Println(err)
			}
			if serverReady == len(cluster) {
				return
			}
		}
	}
}

func WaitForServerToDie(t *testing.T, s ServerInfo) {
	for {
		c := client.CreateColoniesClient("localhost", s.Node.APIPort, true, true)
		err := c.CheckHealth()
		if err != nil {
			return
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func WaitForProcessGraphs(t *testing.T, c *client.ColoniesClient, colonyID string, generatorID string, runtimePrvKey string, threshold int) int {
	var graphs []*core.ProcessGraph
	var err error
	retries := 40
	for i := 0; i < retries; i++ {
		graphs, err = c.GetWaitingProcessGraphs(colonyID, 100, runtimePrvKey)
		assert.Nil(t, err)
		if generatorID != "" {
			c.AddArgToGenerator(generatorID, "arg", runtimePrvKey)
		}
		if len(graphs) > threshold {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return len(graphs)
}
