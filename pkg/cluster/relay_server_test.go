package cluster

import (
	"io/ioutil"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRelayServer(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	nodes := testClusterNodes(t, "etcd1", "etcd2", "etcd3")
	node1, node2, node3 := nodes[0], nodes[1], nodes[2]

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	relayServer1 := CreateRelayServer(node1, config)
	relayServer2 := CreateRelayServer(node2, config)
	relayServer3 := CreateRelayServer(node3, config)

	defer relayServer1.Shutdown()
	defer relayServer2.Shutdown()
	defer relayServer3.Shutdown()

	incomingRelayServer1 := relayServer1.Receive()
	incomingRelayServer2 := relayServer2.Receive()
	incomingRelayServer3 := relayServer3.Receive()

	relayServer1Received := make(map[string]int)
	relayServer2Received := make(map[string]int)
	relayServer3Received := make(map[string]int)

	relayServer1Wait := make(chan struct{})
	relayServer2Wait := make(chan struct{})
	relayServer3Wait := make(chan struct{})

	expectedNrMessage := 2

	go func() {
		counter := 0
		for {
			select {
			case msg := <-incomingRelayServer1:
				if val, ok := relayServer1Received[string(msg.Data)]; ok {
					val++
					relayServer1Received[string(msg.Data)] = val
					counter++
				} else {
					relayServer1Received[string(msg.Data)] = 1
					counter++
				}
				if counter == expectedNrMessage {
					relayServer1Wait <- struct{}{}
				}
			}
		}
	}()
	go func() {
		counter := 0
		for {
			select {
			case msg := <-incomingRelayServer2:
				if val, ok := relayServer2Received[string(msg.Data)]; ok {
					val++
					relayServer2Received[string(msg.Data)] = val
					counter++
				} else {
					relayServer2Received[string(msg.Data)] = 1
					counter++
				}
				if counter == expectedNrMessage {
					relayServer2Wait <- struct{}{}
				}
			}
		}
	}()
	go func() {
		counter := 0
		for {
			select {
			case msg := <-incomingRelayServer3:
				if val, ok := relayServer3Received[string(msg.Data)]; ok {
					val++
					relayServer3Received[string(msg.Data)] = val
					counter++
				} else {
					relayServer3Received[string(msg.Data)] = 1
					counter++
				}
				if counter == expectedNrMessage {
					relayServer3Wait <- struct{}{}
				}
			}
		}
	}()

	err := relayServer1.Broadcast([]byte("relayserver1"))
	assert.Nil(t, err)
	err = relayServer2.Broadcast([]byte("relayserver2"))
	assert.Nil(t, err)
	err = relayServer3.Broadcast([]byte("relayserver3"))
	assert.Nil(t, err)

	<-relayServer1Wait
	<-relayServer2Wait
	<-relayServer3Wait

	assert.Equal(t, relayServer1Received["relayserver2"], 1)
	assert.Equal(t, relayServer1Received["relayserver3"], 1)
	assert.Len(t, relayServer1Received, 2)

	assert.Equal(t, relayServer2Received["relayserver1"], 1)
	assert.Equal(t, relayServer2Received["relayserver3"], 1)
	assert.Len(t, relayServer2Received, 2)

	assert.Equal(t, relayServer3Received["relayserver1"], 1)
	assert.Equal(t, relayServer3Received["relayserver2"], 1)
	assert.Len(t, relayServer2Received, 2)
}
