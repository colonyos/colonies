package cluster

import (
	"io"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRelay(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	node1 := &Node{Name: "etcd1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := &Node{Name: "etcd2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := &Node{Name: "etcd3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := EmptyConfig()
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	clusterManager1 := CreateClusterManager(node1, config, ".")
	clusterManager2 := CreateClusterManager(node2, config, ".")
	clusterManager3 := CreateClusterManager(node3, config, ".")

	clusterManager1.BlockUntilReady()
	clusterManager2.BlockUntilReady()
	clusterManager3.BlockUntilReady()

	relay1 := clusterManager1.Relay()
	relay2 := clusterManager2.Relay()
	relay3 := clusterManager3.Relay()

	defer clusterManager1.Shutdown()
	defer clusterManager2.Shutdown()
	defer clusterManager3.Shutdown()

	incomingRelayChan1 := relay1.Receive()
	incomingRelayChan2 := relay2.Receive()
	incomingRelayChan3 := relay3.Receive()

	relay1Received := make(map[string]int)
	relay2Received := make(map[string]int)
	relay3Received := make(map[string]int)

	relay1Wait := make(chan struct{})
	relay2Wait := make(chan struct{})
	relay3Wait := make(chan struct{})

	expectedNrMessage := 2

	go func() {
		counter := 0
		for {
			select {
			case msg := <-incomingRelayChan1:
				if val, ok := relay1Received[string(msg)]; ok {
					val++
					relay1Received[string(msg)] = val
					counter++
				} else {
					relay1Received[string(msg)] = 1
					counter++
				}
				if counter == expectedNrMessage {
					relay1Wait <- struct{}{}
				}
			}
		}
	}()
	go func() {
		counter := 0
		for {
			select {
			case msg := <-incomingRelayChan2:
				if val, ok := relay2Received[string(msg)]; ok {
					val++
					relay2Received[string(msg)] = val
					counter++
				} else {
					relay2Received[string(msg)] = 1
					counter++
				}
				if counter == expectedNrMessage {
					relay2Wait <- struct{}{}
				}
			}
		}
	}()
	go func() {
		counter := 0
		for {
			select {
			case msg := <-incomingRelayChan3:
				if val, ok := relay3Received[string(msg)]; ok {
					val++
					relay3Received[string(msg)] = val
					counter++
				} else {
					relay3Received[string(msg)] = 1
					counter++
				}
				if counter == expectedNrMessage {
					relay3Wait <- struct{}{}
				}
			}
		}
	}()

	err := relay1.Broadcast([]byte("relayserver1"))
	assert.Nil(t, err)
	err = relay2.Broadcast([]byte("relayserver2"))
	assert.Nil(t, err)
	err = relay3.Broadcast([]byte("relayserver3"))
	assert.Nil(t, err)

	<-relay1Wait
	<-relay2Wait
	<-relay3Wait

	assert.Equal(t, relay1Received["relayserver2"], 1)
	assert.Equal(t, relay1Received["relayserver3"], 1)
	assert.Len(t, relay1Received, 2)

	assert.Equal(t, relay2Received["relayserver1"], 1)
	assert.Equal(t, relay2Received["relayserver3"], 1)
	assert.Len(t, relay2Received, 2)

	assert.Equal(t, relay3Received["relayserver1"], 1)
	assert.Equal(t, relay3Received["relayserver2"], 1)
	assert.Len(t, relay2Received, 2)
}
