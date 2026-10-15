package server

import (
	"context"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/cluster"
	"github.com/colonyos/colonies/pkg/constants"
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

// Retention settings for test servers created with retention enabled. The
// policy must leave enough room for a test to submit, assign, close and
// inspect a process before it is deleted, even under heavy load (for example
// go test -race -count=10 ./...), so it is much longer than the worker period.
const (
	testRetentionPolicySeconds = 5
	testRetentionPeriodMillis  = 500
)

// Cron trigger period for single-node test servers. Production evaluates crons
// once a second, which makes every cron test wait two to three seconds for the
// first trigger. A shorter tick keeps the same semantics (the one second cron
// interval is still the floor) while cutting the waiting to just over a second.
const testCronPeriodMillis = 100

func prepareTestsWithRetention(t *testing.T, retention bool) (*client.ColoniesClient, *Server, string, chan bool) {
	// Dynamic ports and a per-test etcd data directory allow test packages to
	// run in parallel without colliding on fixed ports or /tmp/colonies. The
	// ports stay reserved until just before each component binds, so parallel
	// test processes cannot be handed the same port during the slow parts of
	// the setup (database preparation, etcd startup).
	reserved := utils.ReservePortsOrPanic(4)
	apiPort, etcdClientPort, etcdPeerPort, relayPort := reserved[0].Port(), reserved[1].Port(), reserved[2].Port(), reserved[3].Port()
	etcdDataPath := t.TempDir()

	client := client.CreateColoniesClient(constants.TESTHOST, apiPort, Insecure, SkipTLSVerify)

	db, err := postgresql.PrepareTests()
	assert.Nil(t, err)

	crypto := crypto.CreateCrypto()
	serverPrvKey, err := crypto.GeneratePrivateKey()
	assert.Nil(t, err)
	serverID, err := crypto.GenerateID(serverPrvKey)
	assert.Nil(t, err)

	err = db.SetServerID("", serverID)
	assert.Nil(t, err)

	node := cluster.Node{Name: "etcd", Host: "localhost", EtcdClientPort: etcdClientPort, EtcdPeerPort: etcdPeerPort, RelayPort: relayPort, APIPort: apiPort}
	clusterConfig := cluster.Config{}
	clusterConfig.AddNode(node)
	reserved[1].Release()
	reserved[2].Release()
	reserved[3].Release()
	server := CreateServer(db, apiPort, EnableTLS, "", "", node, clusterConfig, etcdDataPath, constants.GENERATOR_TRIGGER_PERIOD, testCronPeriodMillis, false, false, retention, testRetentionPolicySeconds, testRetentionPeriodMillis, time.Duration(constants.DEFAULT_STALE_EXECUTOR_DURATION)*time.Second)

	done := make(chan bool)
	reserved[0].Release()
	go func() {
		err := server.ServeForever()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			// A silent bind failure would leave the client talking to some
			// other process's server; fail loudly instead
			panic("test server failed to serve: " + err.Error())
		}
		db.Close()
		done <- true
	}()

	waitForServer(t, client)

	return client, server, serverPrvKey, done
}

// waitForServer blocks until the test server answers health checks. The server
// listens in a goroutine, so without this a test on a loaded machine can send
// its first request before the port is bound and fail with a connection error.
func waitForServer(t *testing.T, client *client.ColoniesClient) {
	deadline := time.Now().Add(30 * time.Second)
	var err error
	for time.Now().Before(deadline) {
		err = client.CheckHealth()
		if err == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("test server did not become ready: %v", err)
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
			_, err := server.controller.GetEventHandler().WaitForProcess(process.FunctionSpec.Conditions.ExecutorType, state, process.ID, process.FunctionSpec.Conditions.LocationName, ctx)
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
	return startCluster(t, db, size, true)
}

// StartClusterDistributed creates a cluster with ExclusiveAssign=false for testing distributed assignment
func StartClusterDistributed(t *testing.T, db database.Database, size int) []ServerInfo {
	return startCluster(t, db, size, false)
}

func startCluster(t *testing.T, db database.Database, size int, exclusiveAssign bool) []ServerInfo {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	etcdDataPath := t.TempDir()

	// Ports stay reserved until just before each component binds so parallel
	// test processes cannot be handed the same port while the cluster starts
	reserved := utils.ReservePortsOrPanic(4 * size)
	t.Cleanup(func() { utils.ReleasePorts(reserved) })

	clusterConfig := cluster.Config{}
	for i := 0; i < size; i++ {
		node := cluster.Node{
			Name:           "etcd" + strconv.Itoa(i),
			Host:           "localhost",
			EtcdClientPort: reserved[4*i].Port(),
			EtcdPeerPort:   reserved[4*i+1].Port(),
			RelayPort:      reserved[4*i+2].Port(),
			APIPort:        reserved[4*i+3].Port()}
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
			// CreateServer binds the etcd and relay ports
			reserved[4*i].Release()
			reserved[4*i+1].Release()
			reserved[4*i+2].Release()
			server := CreateServer(db, node.APIPort, false, "", "", node, clusterConfig, etcdDataPath+"/etcd"+strconv.Itoa(i), constants.GENERATOR_TRIGGER_PERIOD, constants.CRON_TRIGGER_PERIOD, exclusiveAssign, false, false, -1, 500, time.Duration(constants.DEFAULT_STALE_EXECUTOR_DURATION)*time.Second)
			done := make(chan struct{})
			s := ServerInfo{ServerID: serverID, ServerPrvKey: serverPrvKey, Server: server, Node: node, Done: done}
			reserved[4*i+3].Release()
			go func(i int) {
				log.Info("ColoniesServer serving")
				err := server.ServeForever()
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					// A silent bind failure would let WaitForCluster count
					// another process's server as ours; fail loudly instead
					panic("cluster server " + strconv.Itoa(i) + " failed to serve: " + err.Error())
				}
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

// WaitForCluster blocks until every server in the cluster answers health
// checks in the same pass, so one healthy node polled repeatedly cannot
// satisfy the wait on behalf of nodes that are still starting.
func WaitForCluster(t *testing.T, cluster []ServerInfo) {
	for {
		serverReady := 0
		for _, s := range cluster {
			client := client.CreateColoniesClient("localhost", s.Node.APIPort, true, true)
			if err := client.CheckHealth(); err == nil {
				serverReady++
			}
		}
		if serverReady == len(cluster) {
			return
		}
		time.Sleep(50 * time.Millisecond)
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

// transientProcessGraphError is the message the server returns when a graph is
// visible before all of its processes are. CreateProcessGraph inserts the graph
// first, so a poll can briefly land in that window; the assignment path retries
// graph resolution for the same reason.
const transientProcessGraphError = "Failed to iterate processgraph, process with ID="

// WaitForProcessGraphs polls until at least threshold waiting process graphs
// exist or the deadline passes, and returns the number of graphs seen on the
// last poll. The transient missing-process error is retried until the deadline;
// any other error fails the test immediately.
func WaitForProcessGraphs(t *testing.T, c *client.ColoniesClient, colonyName string, generatorID string, executorPrvKey string, threshold int) int {
	var graphs []*core.ProcessGraph
	var err error
	deadline := time.Now().Add(40 * time.Second)
	for {
		graphs, err = c.GetWaitingProcessGraphs(colonyName, 100, executorPrvKey)
		if err != nil && !strings.Contains(err.Error(), transientProcessGraphError) {
			assert.Nil(t, err)
			break
		}
		if err == nil && len(graphs) >= threshold {
			break
		}
		if time.Now().After(deadline) {
			assert.Nil(t, err, "process graphs still not consistent at deadline")
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	return len(graphs)
}
