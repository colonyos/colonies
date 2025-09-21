package server

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/constants"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/database"
	"github.com/colonyos/colonies/pkg/rpc"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/colonies/pkg/server/controllers"
	"github.com/colonyos/colonies/pkg/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type TestEnv1 struct {
	Colony1PrvKey   string
	Colony1Name     string
	Colony1ID       string
	Colony2PrvKey   string
	Colony2Name     string
	Colony2ID       string
	Executor1PrvKey string
	Executor1Name   string
	Executor1ID     string
	Executor2PrvKey string
	Executor2Name   string
	Executor2ID     string
}

type testEnv1 struct {
	colony1PrvKey   string
	colony1Name     string
	colony1ID       string
	colony2PrvKey   string
	colony2Name     string
	colony2ID       string
	executor1PrvKey string
	executor1Name   string
	executor1ID     string
	executor2PrvKey string
	executor2Name   string
	executor2ID     string
}

type TestEnv2 struct {
	ColonyID       string
	ColonyName     string
	Colony         *core.Colony
	ColonyPrvKey   string
	ExecutorName   string
	ExecutorID     string
	Executor       *core.Executor
	ExecutorPrvKey string
}

type testEnv2 struct {
	colonyID       string
	colonyName     string
	colony         *core.Colony
	colonyPrvKey   string
	executorName   string
	executorID     string
	executor       *core.Executor
	executorPrvKey string
}

const EnableTLS = false
const Insecure = true
const SkipTLSVerify = false

func SetupTestEnv1(t *testing.T) (*TestEnv1, *client.ColoniesClient, *Server, string, chan bool) {
	env, client, server, serverPrvKey, done := setupTestEnv1(t)
	return &TestEnv1{
		Colony1PrvKey:   env.colony1PrvKey,
		Colony1Name:     env.colony1Name,
		Colony1ID:       env.colony1ID,
		Colony2PrvKey:   env.colony2PrvKey,
		Colony2Name:     env.colony2Name,
		Colony2ID:       env.colony2ID,
		Executor1PrvKey: env.executor1PrvKey,
		Executor1Name:   env.executor1Name,
		Executor1ID:     env.executor1ID,
		Executor2PrvKey: env.executor2PrvKey,
		Executor2Name:   env.executor2Name,
		Executor2ID:     env.executor2ID,
	}, client, server, serverPrvKey, done
}

func setupTestEnv1(t *testing.T) (*testEnv1, *client.ColoniesClient, *Server, string, chan bool) {
	rand.Seed(time.Now().UTC().UnixNano())

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	//log.SetLevel(log.DebugLevel)

	client, server, serverPrvKey, done := prepareTests(t)

	colony1, colony1PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony1, serverPrvKey)
	assert.Nil(t, err)

	colony2, colony2PrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony2, serverPrvKey)
	assert.Nil(t, err)

	executor1, executor1PrvKey, err := utils.CreateTestExecutorWithKey(colony1.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor1, colony1PrvKey)
	assert.Nil(t, err)

	executor2, executor2PrvKey, err := utils.CreateTestExecutorWithKey(colony2.Name)
	assert.Nil(t, err)
	_, err = client.AddExecutor(executor2, colony2PrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(colony1.Name, executor1.Name, colony1PrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(colony2.Name, executor2.Name, colony2PrvKey)
	assert.Nil(t, err)

	env := &testEnv1{colony1PrvKey: colony1PrvKey,
		colony1Name:     colony1.Name,
		colony1ID:       colony1.ID,
		colony2PrvKey:   colony2PrvKey,
		colony2Name:     colony2.Name,
		colony2ID:       colony2.ID,
		executor1PrvKey: executor1PrvKey,
		executor1ID:     executor1.ID,
		executor1Name:   executor1.Name,
		executor2PrvKey: executor2PrvKey,
		executor2ID:     executor2.ID,
		executor2Name:   executor2.Name}

	return env, client, server, serverPrvKey, done
}

func setupTestEnv2(t *testing.T) (*testEnv2, *client.ColoniesClient, *Server, string, chan bool) {
	rand.Seed(time.Now().UTC().UnixNano())

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard
	//log.SetLevel(log.DebugLevel)
	client, server, serverPrvKey, done := prepareTests(t)

	colony, colonyPrvKey, err := utils.CreateTestColonyWithKey()
	assert.Nil(t, err)
	_, err = client.AddColony(colony, serverPrvKey)
	assert.Nil(t, err)

	executor, executorPrvKey, err := utils.CreateTestExecutorWithKey(colony.Name)
	_, err = client.AddExecutor(executor, colonyPrvKey)
	assert.Nil(t, err)

	err = client.ApproveExecutor(colony.Name, executor.Name, colonyPrvKey)
	assert.Nil(t, err)

	env := &testEnv2{
		colonyID:       colony.ID,
		colonyName:     colony.Name,
		colony:         colony,
		colonyPrvKey:   colonyPrvKey,
		executorName:   executor.Name,
		executorID:     executor.ID,
		executor:       executor,
		executorPrvKey: executorPrvKey}

	return env, client, server, serverPrvKey, done
}

func SetupTestEnv2(t *testing.T) (*TestEnv2, *client.ColoniesClient, *Server, string, chan bool) {
	env, client, server, serverPrvKey, done := setupTestEnv2(t)
	return &TestEnv2{
		ColonyID:       env.colonyID,
		ColonyName:     env.colonyName,
		Colony:         env.colony,
		ColonyPrvKey:   env.colonyPrvKey,
		ExecutorName:   env.executorName,
		ExecutorID:     env.executorID,
		Executor:       env.executor,
		ExecutorPrvKey: env.executorPrvKey,
	}, client, server, serverPrvKey, done
}

func PrepareTests(t *testing.T) (*client.ColoniesClient, *Server, string, chan bool) {
	return prepareTests(t)
}

func prepareTests(t *testing.T) (*client.ColoniesClient, *Server, string, chan bool) {
	return prepareTestsWithRetention(t, false)
}

func prepareTestsWithRetention(t *testing.T, retention bool) (*client.ColoniesClient, *Server, string, chan bool) {
	os.RemoveAll("/tmp/colonies")
	client := client.CreateColoniesClient(constants.TESTHOST, constants.TESTPORT, Insecure, SkipTLSVerify)

	db, err := database.PrepareTests()
	assert.Nil(t, err)

	crypto := crypto.CreateCrypto()
	serverPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := crypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)

	err = db.SetServerID("", serverID)
	assert.Nil(t, err)

	node := cluster.Node{Name: "etcd", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: constants.TESTPORT}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	server := CreateServerFromEnv(db, constants.TESTPORT, EnableTLS, "", "", node, clusterConfig, "/tmp/colonies/etcd", constants.GENERATOR_TRIGGER_PERIOD, constants.CRON_TRIGGER_PERIOD, true, false, retention, 1, 500)

	done := make(chan bool)
	go func() {
		server.ServeForever()
		db.Close()
		done <- true
	}()

	return client, server, serverPrvKey, done
}

func createTestColoniesController(db database.Database) *controllers.ColoniesController {
	node := cluster.Node{Name: "etcd", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: constants.TESTPORT}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	return controllers.CreateColoniesController(db, node, clusterConfig, "/tmp/colonies/etcd", constants.GENERATOR_TRIGGER_PERIOD, constants.CRON_TRIGGER_PERIOD, false, -1, 500)
}

func createTestColoniesController2(db database.Database) *controllers.ColoniesController {
	node := cluster.Node{Name: "etcd2", Host: "localhost", EtcdClientPort: 26100, EtcdPeerPort: 27100, RelayPort: 28100, APIPort: constants.TESTPORT}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	return controllers.CreateColoniesController(db, node, clusterConfig, "/tmp/colonies/etcd", constants.GENERATOR_TRIGGER_PERIOD, constants.CRON_TRIGGER_PERIOD, false, -1, 500)
}

func GenerateDiamondtWorkflowSpec(colonyName string) *core.WorkflowSpec {
	//         task1
	//          / \
	//     task2   task3
	//          \ /
	//         task4

	workflowSpec := core.CreateWorkflowSpec(colonyName)

	funcSpec1 := core.CreateEmptyFunctionSpec()
	funcSpec1.NodeName = "task1"
	funcSpec1.Conditions.ColonyName = colonyName
	funcSpec1.Conditions.ExecutorType = "test_executor_type"

	funcSpec2 := core.CreateEmptyFunctionSpec()
	funcSpec2.NodeName = "task2"
	funcSpec2.Conditions.ColonyName = colonyName
	funcSpec2.Conditions.ExecutorType = "test_executor_type"

	funcSpec3 := core.CreateEmptyFunctionSpec()
	funcSpec3.NodeName = "task3"
	funcSpec3.Conditions.ColonyName = colonyName
	funcSpec3.Conditions.ExecutorType = "test_executor_type"

	funcSpec4 := core.CreateEmptyFunctionSpec()
	funcSpec4.NodeName = "task4"
	funcSpec4.Conditions.ColonyName = colonyName
	funcSpec4.Conditions.ExecutorType = "test_executor_type"

	funcSpec2.AddDependency("task1")
	funcSpec3.AddDependency("task1")
	funcSpec4.AddDependency("task2")
	funcSpec4.AddDependency("task3")

	workflowSpec.AddFunctionSpec(funcSpec1)
	workflowSpec.AddFunctionSpec(funcSpec2)
	workflowSpec.AddFunctionSpec(funcSpec3)
	workflowSpec.AddFunctionSpec(funcSpec4)

	return workflowSpec
}

func GenerateTreeWorkflowSpec(colonyName string) *core.WorkflowSpec {
	//         task1
	//          / \
	//     task2   task3

	workflowSpec := core.CreateWorkflowSpec(colonyName)

	funcSpec1 := core.CreateEmptyFunctionSpec()
	funcSpec1.NodeName = "task1"
	funcSpec1.Conditions.ColonyName = colonyName
	funcSpec1.Conditions.ExecutorType = "test_executor_type"

	funcSpec2 := core.CreateEmptyFunctionSpec()
	funcSpec2.NodeName = "task2"
	funcSpec2.Conditions.ColonyName = colonyName
	funcSpec2.Conditions.ExecutorType = "test_executor_type"

	funcSpec3 := core.CreateEmptyFunctionSpec()
	funcSpec3.NodeName = "task3"
	funcSpec3.Conditions.ColonyName = colonyName
	funcSpec3.Conditions.ExecutorType = "test_executor_type"

	funcSpec2.AddDependency("task1")
	funcSpec3.AddDependency("task1")

	workflowSpec.AddFunctionSpec(funcSpec1)
	workflowSpec.AddFunctionSpec(funcSpec2)
	workflowSpec.AddFunctionSpec(funcSpec3)

	return workflowSpec
}

func GenerateSingleWorkflowSpec(colonyName string) *core.WorkflowSpec {
	workflowSpec := core.CreateWorkflowSpec(colonyName)
	funcSpec1 := core.CreateEmptyFunctionSpec()
	funcSpec1.NodeName = "task1"
	funcSpec1.Conditions.ColonyName = colonyName
	funcSpec1.Conditions.ExecutorType = "test_executor_type"

	workflowSpec.AddFunctionSpec(funcSpec1)

	return workflowSpec
}

func WaitForProcesses(t *testing.T, server *Server, processes []*core.Process, state int) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancelCtx()
	wait := make(chan error)
	for _, process := range processes {
		go func(process *core.Process) {
			_, err := server.controller.GetEventHandler().WaitForProcess(process.FunctionSpec.Conditions.ExecutorType, state, process.ID, ctx)
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
	Server       *Server
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

	db.SetServerID("", serverID)

	sChan := make(chan ServerInfo)
	for i, node := range clusterConfig.Nodes {
		go func(i int, node cluster.Node) {
			log.WithFields(log.Fields{"APIPort": node.APIPort}).Info("Starting ColoniesServer")
			server := CreateServerFromEnv(db, node.APIPort, false, "", "", node, clusterConfig, "/tmp/colonies/etcd"+strconv.Itoa(i), constants.GENERATOR_TRIGGER_PERIOD, constants.CRON_TRIGGER_PERIOD, true, false, false, -1, 500)
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

func WaitForProcessGraphs(t *testing.T, c *client.ColoniesClient, colonyName string, generatorID string, executorPrvKey string, threshold int) int {
	var graphs []*core.ProcessGraph
	var err error
	retries := 40
	for i := 0; i < retries; i++ {
		graphs, err = c.GetWaitingProcessGraphs(colonyName, 100, executorPrvKey)
		assert.Nil(t, err)
		if len(graphs) >= threshold {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return len(graphs)
}
