package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelayServer(t *testing.T) {
	node1 := Node{Name: "etcd1", Host: "localhost", EtcdClientPort: 24100, EtcdPeerPort: 23100, RelayPort: 25100, APIPort: 26100}
	node2 := Node{Name: "etcd2", Host: "localhost", EtcdClientPort: 24200, EtcdPeerPort: 23200, RelayPort: 25200, APIPort: 26200}
	node3 := Node{Name: "etcd3", Host: "localhost", EtcdClientPort: 24300, EtcdPeerPort: 23300, RelayPort: 25300, APIPort: 26300}

	config := Config{}
	config.AddNode(node1)
	config.AddNode(node2)
	config.AddNode(node3)

	relayServer1 := CreateRelayServer(node1, config)
	relayServer2 := CreateRelayServer(node2, config)
	relayServer3 := CreateRelayServer(node3, config)

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
				if val, ok := relayServer1Received[string(msg)]; ok {
					val++
					relayServer1Received[string(msg)] = val
					counter++
				} else {
					relayServer1Received[string(msg)] = 1
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
				if val, ok := relayServer2Received[string(msg)]; ok {
					val++
					relayServer2Received[string(msg)] = val
					counter++
				} else {
					relayServer2Received[string(msg)] = 1
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
				if val, ok := relayServer3Received[string(msg)]; ok {
					val++
					relayServer3Received[string(msg)] = val
					counter++
				} else {
					relayServer3Received[string(msg)] = 1
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
