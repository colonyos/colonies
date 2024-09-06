package cluster

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const PING_RESPONSE_TIMEOUT = 1 // Wait max 1 second for a response to a ping request
const RPC_PURGE_INTERVAL = 600  // Purge RPC responses every 10 minutes
const NODE_LIST_RETRY_DELAY = PING_RESPONSE_TIMEOUT + 1

type Coordinator struct {
	thisNode          Node
	clusterConfig     Config
	etcdServer        *EtcdServer
	rpc               *clusterRPC
	doneChan          chan bool
	readyChan         chan bool
	nodeList          []string
	nodeListChecksum  string
	nodeListMutex     *sync.Mutex
	simFailureMutex   *sync.Mutex
	genListInProgress bool
	genListDoneChan   chan bool
	simulateFailures  bool
	failureTimeout    time.Duration
}

func CreateCoordinator(thisNode Node, clusterConfig Config, etcdServer *EtcdServer, ginHandler *gin.Engine) *Coordinator {
	c := &Coordinator{
		thisNode:        thisNode,
		clusterConfig:   clusterConfig,
		etcdServer:      etcdServer,
		rpc:             createClusterRPC(thisNode, clusterConfig, ginHandler, time.Duration(time.Second*RPC_PURGE_INTERVAL)),
		doneChan:        make(chan bool),
		readyChan:       make(chan bool),
		nodeListMutex:   &sync.Mutex{},
		simFailureMutex: &sync.Mutex{},
	}

	go c.handleRequests()

	<-c.readyChan

	return c
}

func (c *Coordinator) EnableFailures(failureTimeout time.Duration) {
	c.simFailureMutex.Lock()
	defer c.simFailureMutex.Unlock()
	log.WithFields(log.Fields{"Node": c.thisNode.Name, "Timeout": failureTimeout}).Debug("Enabling failures")
	c.simulateFailures = true
	c.failureTimeout = failureTimeout
}

func (c *Coordinator) DisableFailures() {
	c.simFailureMutex.Lock()
	defer c.simFailureMutex.Unlock()
	c.simulateFailures = false
}

func (c *Coordinator) handleRequests() {
	log.WithFields(log.Fields{"Node": c.thisNode.Name}).Debug("Handling requests")

	close(c.readyChan)

	msgChan := c.rpc.receiveChan()
	for {
		select {
		case msg := <-msgChan:
			c.simFailureMutex.Lock()
			if c.simulateFailures {
				c.simFailureMutex.Unlock()
				log.WithFields(log.Fields{"Node": c.thisNode.Name, "SleepTime": c.failureTimeout}).Debug("Simulating failure")
				time.Sleep(c.failureTimeout)
			} else {
				c.simFailureMutex.Unlock()
			}

			log.WithFields(log.Fields{"Node": c.thisNode.Name, "MsgType": msg.MsgType}).Debug("Received message")
			switch msg.MsgType {
			case PingRequest:
				c.handlePingRequest(msg)
			case VerifyNodeListRequest:
				c.handleVerifyNodeListRequest(msg)
			case NodeListRequest:
				c.handleGetNodeListRequest(msg)
			case RPCRequest:
				c.handleRPCRequest(msg)
			}
		case <-c.doneChan:
			log.WithFields(log.Fields{"Node": c.thisNode.Name}).Debug("Closing coordinator")
			return
		}
	}
}

func (c *Coordinator) handlePingRequest(msg *ClusterMsg) {
	log.Debugf("Received Ping request from %s", msg.Originator)
	msg.MsgType = PingResponse
	c.rpc.reply(msg)
}

func (c *Coordinator) handleVerifyNodeListRequest(msg *ClusterMsg) {
	log.Debugf("Received VerifyNodeList request from %s", msg.Originator)

}

func (c *Coordinator) handleGetNodeListRequest(msg *ClusterMsg) {
}

func (c *Coordinator) handleRPCRequest(msg *ClusterMsg) {
	log.Debugf("Received RPC request from %s", msg.Originator)
}

func (c *Coordinator) calcChecksum(nodeList []string) string {
	concatenated := strings.Join(nodeList, ",")
	hash := sha256.New()
	hash.Write([]byte(concatenated))
	return hex.EncodeToString(hash.Sum(nil))
}

func (c *Coordinator) GetNodeList() []string {
	c.nodeListMutex.Lock()
	defer c.nodeListMutex.Unlock()

	return c.nodeList
}

func (c *Coordinator) LatestNodeListVersion(checksum string) bool {
	c.nodeListMutex.Lock()
	defer c.nodeListMutex.Unlock()
	return c.nodeListChecksum == checksum
}

func (c *Coordinator) LeaderName() string {
	return c.etcdServer.Leader()
}

func (c *Coordinator) LeaderNode() Node {
	return c.etcdServer.CurrentCluster().Leader
}

func (c *Coordinator) IsLeader() bool {
	return c.etcdServer.Leader() == c.thisNode.Name
}

func (c *Coordinator) genNodeList() bool {
	c.nodeListMutex.Lock()
	if c.genListInProgress {
		log.WithFields(log.Fields{"Node": c.thisNode.Name}).Debug("Node list generation already in progress, waiting for it to finish")
		c.nodeListMutex.Unlock()

		// Wait for the current list node generation to finish and ruturn the last generated list
		<-c.genListDoneChan
		return false
	}

	c.genListInProgress = true
	c.genListDoneChan = make(chan bool)
	c.nodeListMutex.Unlock()

	responsesChan := make(chan *response, len(c.rpc.clusterConfig.Nodes)-1)

	log.WithFields(log.Fields{"Node": c.thisNode.Name}).Debug("Sending ping requests to all nodes")

	for _, node := range c.rpc.clusterConfig.Nodes {
		if node.Name != c.thisNode.Name {
			msg := &ClusterMsg{
				MsgType:   PingRequest,
				Recipient: node.Name,
			}

			log.WithFields(log.Fields{"Sender": c.thisNode.Name, "Reciever": node.Name}).Debug("Sending ping request")
			response, err := c.rpc.sendAndReceive(node.Name, msg)
			if err != nil {
				log.WithFields(log.Fields{"error": err}).Error("Failed to send ping request")
				continue
			}
			responsesChan <- response
		}
	}
	close(responsesChan)

	pingResponses := make(chan *ClusterMsg, len(c.rpc.clusterConfig.Nodes))
	var wg sync.WaitGroup // Use a WaitGroup to wait for all goroutines to finish

	//for replyChan := range replyChans {
	for resp := range responsesChan {
		wg.Add(1)
		go func(resp *response) {
			replyChan := resp.receiveChan
			errChan := resp.errChan
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
			defer cancel()
			defer c.rpc.close(resp)

			select {
			case <-ctx.Done():
				log.WithFields(log.Fields{"Node": c.thisNode.Name}).Debug("Timeout waiting for ping response")
			case err := <-errChan:
				log.WithFields(log.Fields{"Error": err}).Error("Error waiting for ping response")
			case replyMsg := <-replyChan:
				if replyMsg != nil { // Ensure we're not processing nil messages
					if replyMsg.MsgType == PingResponse {
						log.Debugf("Received ping response from %s", replyMsg.Originator)
					} else {
						log.WithFields(log.Fields{"msgType": replyMsg.MsgType}).Error("Unexpected message type, expected ping response")
					}
					pingResponses <- replyMsg
				}
			}

		}(resp)
	}

	log.WithFields(log.Fields{"Node": c.thisNode.Name}).Debug("Waiting for all ping responses to generate node list")
	wg.Wait()
	close(pingResponses) // Close pingResponses when all goroutines are done

	log.WithFields(log.Fields{"Node": c.thisNode.Name}).Debug("Generating node list")

	nodeList := make([]string, 0)
	for msg := range pingResponses {
		nodeList = append(nodeList, msg.Originator)
	}

	nodeList = append(nodeList, c.thisNode.Name)

	c.nodeListMutex.Lock()
	c.genListInProgress = false
	c.nodeList = nodeList
	c.nodeListChecksum = c.calcChecksum(nodeList)
	c.nodeListMutex.Unlock()

	close(c.genListDoneChan)

	log.WithFields(log.Fields{"Node": c.thisNode.Name, "NodeList": nodeList}).Debug("Done generating node list")

	// If the nodelist contains less nodes than the cluster configuration,
	// spawn a thread to tries to connect to the missing nodes
	if c.IsLeader() {
		if len(nodeList) < len(c.clusterConfig.Nodes) {
			go func() {
				for {

					log.WithFields(log.Fields{"Node": c.thisNode.Name}).Debug("Some nodes are missing from the node list, try to generate a new one")
					time.Sleep((NODE_LIST_RETRY_DELAY) * time.Second)
					if c.genNodeList() {
						break
					}
				}
			}()
		}
	}

	return true
}

func (c *Coordinator) shutdown() {
	close(c.doneChan)
	c.rpc.shutdown()
}
