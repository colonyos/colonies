package mock

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/colonyos/colonies/pkg/p2p"
	"github.com/stretchr/testify/assert"
)

func TestMessengerSendAndForget(t *testing.T) {
	n := CreateFakeNetwork()

	node1 := p2p.Node{Addr: "10.0.0.1:1111"}
	messenger1 := CreateMessenger(n, node1)

	node2 := p2p.Node{Addr: "10.0.0.2:1111"}
	messenger2 := CreateMessenger(n, node2)

	msgChan := make(chan p2p.Message)
	ctx := context.TODO()
	go func() {
		messenger1.ListenForever(msgChan, ctx)
	}()

	for {
		err := messenger2.SendAndForget(p2p.Message{From: node2, To: node1, Payload: []byte("Hello")}, context.TODO())
		if err != nil {
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}

	msg := <-msgChan

	assert.Equal(t, string(msg.Payload), "Hello")
	assert.Equal(t, msg.From.Addr, "10.0.0.2:1111")
	assert.Equal(t, msg.To.Addr, "10.0.0.1:1111")

	err := messenger2.SendAndForget(p2p.Message{From: node2, To: node1, Payload: []byte("Hello 2")}, context.TODO())
	assert.Equal(t, err, nil)

	msg = <-msgChan
	assert.Equal(t, string(msg.Payload), "Hello 2")
}

func TestMessengerSendWithReply(t *testing.T) {
	n := CreateFakeNetwork()

	node1 := p2p.Node{Addr: "10.0.0.1:1111"}
	messenger1 := CreateMessenger(n, node1)

	node2 := p2p.Node{Addr: "10.0.0.2:1111"}
	messenger2 := CreateMessenger(n, node2)

	msgChan := make(chan p2p.Message)
	ctx := context.TODO()
	go func() {
		messenger1.ListenForever(msgChan, ctx)
	}()

	replyChan := make(chan *p2p.Message)
	for {
		err := messenger2.SendWithReply(p2p.Message{From: node2, To: node1, Payload: []byte("Hello")}, replyChan, context.TODO())
		if err != nil {
			time.Sleep(10 * time.Millisecond)
		} else {
			break
		}
	}

	fmt.Println("waiting for reply 11111")
	msg := <-msgChan
	fmt.Println(msg)

	assert.Equal(t, string(msg.Payload), "Hello")
	assert.Equal(t, msg.From.Addr, "10.0.0.2:1111")
	assert.Equal(t, msg.To.Addr, "10.0.0.1:1111")

	fmt.Println("sending reply")
	err := messenger1.Reply(msg, "Hi", ctx)
	assert.Nil(t, err)
	fmt.Println("sending reply: done")

	fmt.Println("---------------------------------- 1111")
	msg2 := <-replyChan
	fmt.Println("waiting for reply 11111 2")
	fmt.Println(msg2)
}
