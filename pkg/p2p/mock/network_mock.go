package mock

import (
	"errors"
	"fmt"
	"sync"

	"github.com/colonyos/colonies/pkg/p2p"
)

type FakeNetwork struct {
	Hosts map[string]*FakeSocket
	mutex sync.Mutex
}

func CreateFakeNetwork() *FakeNetwork {
	return &FakeNetwork{Hosts: make(map[string]*FakeSocket)}
}

func (n *FakeNetwork) Listen(addr string) (Socket, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	socket := &FakeSocket{conn: make(chan p2p.Message, 1000)}
	n.Hosts[addr] = socket
	return socket, nil
}

func (n *FakeNetwork) Dial(addr string) (Socket, error) {
	fmt.Println("Dialing:", addr)
	for hostAddr, host := range n.Hosts {
		fmt.Println(hostAddr, host)
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()
	if _, ok := n.Hosts[addr]; !ok {
		return nil, errors.New("No such host " + addr)
	}
	return n.Hosts[addr], nil
}
