package cluster

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestClusterSend(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	node1 := Node{Name: "replica1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := Node{Name: "replica2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := Node{Name: "replica3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	clusterManager1 := CreateClusterManager(node1, config, ".")
	clusterManager2 := CreateClusterManager(node2, config, ".")
	clusterManager3 := CreateClusterManager(node3, config, ".")

	clusterManager1.BlockUntilReady()
	clusterManager2.BlockUntilReady()
	clusterManager3.BlockUntilReady()

	clusterReplica1 := clusterManager1.Cluster()
	clusterReplica2 := clusterManager2.Cluster()

	defer clusterManager1.Shutdown()
	defer clusterManager2.Shutdown()
	defer clusterManager3.Shutdown()

	incomingChan2 := clusterReplica2.ReceiveChan()

	clusterReplica2Wait := make(chan struct{})

	var errReplica2 error

	go func() {
		for {
			select {
			case msg := <-incomingChan2:
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

	err := clusterReplica1.Send("replica_does_not_exists", msg, context.TODO())
	assert.NotNil(t, err)

	err = clusterReplica1.Send("replica2", nil, context.TODO())
	assert.NotNil(t, err)

	err = clusterReplica1.Send("replica2", msg, context.TODO())
	assert.Nil(t, err)

	<-clusterReplica2Wait

	assert.Nil(t, errReplica2)
}

func TestClusterSendAndReceive(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	node1 := Node{Name: "replica1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := Node{Name: "replica2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)

	clusterManager1 := CreateClusterManager(node1, config, ".")
	clusterManager2 := CreateClusterManager(node2, config, ".")

	clusterReplica1 := clusterManager1.Cluster()
	clusterReplica2 := clusterManager2.Cluster()

	defer clusterManager1.Shutdown()
	defer clusterManager2.Shutdown()

	incomingChan2 := clusterReplica2.ReceiveChan()

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
				errSendReplica2 = clusterReplica2.Reply(msg, context.TODO())
				doneReplica2 <- struct{}{}
			}
		}
	}()

	msg := &ClusterMsg{
		MsgType:   PingRequest,
		Recipient: "replica1",
		Data:      []byte("ping"),
	}

	replyChan, err := clusterReplica1.SendAndReceive("replica2", nil, context.TODO())
	assert.NotNil(t, err)

	replyChan, err = clusterReplica1.SendAndReceive("replica_does_not_exists", msg, context.TODO())
	assert.NotNil(t, err)

	replyChan, err = clusterReplica1.SendAndReceive("replica2", msg, context.TODO())
	assert.Nil(t, err)

	reply := <-replyChan
	assert.NotNil(t, reply)
	assert.Equal(t, PingResponse, reply.MsgType)
	assert.Equal(t, "pong", string(reply.Data))

	close(replyChan)

	<-doneReplica2

	assert.Nil(t, errMsgReplica2)
	assert.Nil(t, errSendReplica2)
}
