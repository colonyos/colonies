package cluster

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestClusterRPCSend(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	log.SetLevel(log.DebugLevel)

	node1 := Node{Name: "replica1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := Node{Name: "replica2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := Node{Name: "replica3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	testServer1 := createTestRPCServer(node1, config, ".", time.Duration(10*time.Second))
	testServer2 := createTestRPCServer(node2, config, ".", time.Duration(10*time.Second))
	testServer3 := createTestRPCServer(node3, config, ".", time.Duration(10*time.Second))

	rpcClusterReplica1 := testServer1.clusterRPC()
	rpcClusterReplica2 := testServer2.clusterRPC()

	defer testServer1.Shutdown()
	defer testServer2.Shutdown()
	defer testServer3.Shutdown()

	incomingChan2 := rpcClusterReplica2.receiveChan()

	clusterReplica2Wait := make(chan struct{})

	var errReplica2 error

	go func() {
		for {
			fmt.Println("Waiting for message from replica1")
			select {
			case msg := <-incomingChan2:
				fmt.Println("Received message from replica1")
				if msg == nil {
					errReplica2 = fmt.Errorf("msg is nil")
				}
				if msg != nil && string(msg.Data) != "payload" {
					errReplica2 = fmt.Errorf("invalid payload")
				}
				clusterReplica2Wait <- struct{}{}
			}
		}
	}()

	msg := &ClusterMsg{
		MsgType:   PingRequest,
		Recipient: "replica1",
		Data:      []byte("payload"),
	}

	err := rpcClusterReplica1.send("replica_does_not_exists", msg, context.TODO())
	assert.NotNil(t, err)

	err = rpcClusterReplica1.send("replica2", nil, context.TODO())
	assert.NotNil(t, err)

	fmt.Println("Sending message to replica2")

	err = rpcClusterReplica1.send("replica2", msg, context.TODO())
	assert.Nil(t, err)

	fmt.Println("Sending message to replica2 done")

	<-clusterReplica2Wait

	assert.Nil(t, errReplica2)
}

func TestClusterRPCSendAndReceive(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	node1 := Node{Name: "replica1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := Node{Name: "replica2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)

	testServer1 := createTestRPCServer(node1, config, ".", time.Duration(10*time.Second))
	testServer2 := createTestRPCServer(node2, config, ".", time.Duration(10*time.Second))

	rpcClusterReplica1 := testServer1.clusterRPC()
	rpcClusterReplica2 := testServer2.clusterRPC()

	defer testServer1.Shutdown()
	defer testServer2.Shutdown()

	incomingChan2 := rpcClusterReplica2.receiveChan()

	var errMsgReplica2 error
	var errSendReplica2 error
	var doneReplica2 chan struct{}
	doneReplica2 = make(chan struct{})

	go func() {
		for {
			select {
			case msg := <-incomingChan2:
				if msg == nil {
					errMsgReplica2 = fmt.Errorf("msg is nil")
					break
				}
				payload := string(msg.Data)
				if payload != "ping" {
					errMsgReplica2 = fmt.Errorf("invalid payload")
				}
				msg.Data = []byte("pong")
				msg.MsgType = PingResponse
				errSendReplica2 = rpcClusterReplica2.reply(msg, context.TODO())
				doneReplica2 <- struct{}{}
			}
		}
	}()

	msg := &ClusterMsg{
		MsgType:   PingRequest,
		Recipient: "replica1",
		Data:      []byte("ping"),
	}

	response, err := rpcClusterReplica1.sendAndReceive("replica2", nil, context.TODO())
	defer rpcClusterReplica1.close(response)
	assert.NotNil(t, err)

	response, err = rpcClusterReplica1.sendAndReceive("replica_does_not_exists", msg, context.TODO())
	defer rpcClusterReplica1.close(response)
	assert.NotNil(t, err)

	response, err = rpcClusterReplica1.sendAndReceive("replica2", msg, context.TODO())
	defer rpcClusterReplica1.close(response)
	assert.Nil(t, err)

	reply := <-response.receiveChan
	assert.NotNil(t, reply)
	assert.Equal(t, PingResponse, reply.MsgType)
	assert.Equal(t, "pong", string(reply.Data))

	<-doneReplica2

	assert.Nil(t, errMsgReplica2)
	assert.Nil(t, errSendReplica2)
}

func TestClusterRPCPurge(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard
	log.SetLevel(log.DebugLevel)

	node1 := Node{Name: "replica1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := Node{Name: "replica2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := Node{Name: "replica3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	testServer1 := createTestRPCServer(node1, config, ".", time.Duration(1*time.Second))

	rpc := testServer1.clusterRPC()

	currentTime := time.Now().Unix()

	// Add old and new responses
	rpc.mutex.Lock()
	rpc.pendingResponses["oldMsg"] = &response{
		receiveChan: make(chan *ClusterMsg),
		msgID:       "oldMsg",
		added:       currentTime - 10, // Old
	}
	rpc.pendingResponses["newMsg"] = &response{
		receiveChan: make(chan *ClusterMsg),
		msgID:       "newMsg",
		added:       currentTime + 10, // New
	}

	responses := len(rpc.pendingResponses)
	assert.Equal(t, 2, responses)

	rpc.mutex.Unlock()

	time.Sleep(2 * time.Second)

	rpc.mutex.Lock()
	responses = len(rpc.pendingResponses)
	assert.Equal(t, 1, responses)
	rpc.mutex.Unlock()

	defer testServer1.Shutdown()
}
