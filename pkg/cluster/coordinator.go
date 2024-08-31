package cluster

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const PING_RESPONSE_TIMEOUT = 1 // Wait max 1 second for a response to a ping request

type Coordinator struct {
	coordinator *Coordinator
	etcdServer  *EtcdServer
	cluster     *Cluster
	doneChan    chan bool
}

func CreateCoordinator(etcdServer *EtcdServer, cluster *Cluster) *Coordinator {
	return &Coordinator{etcdServer: etcdServer, cluster: cluster, doneChan: make(chan bool)}
}

func (c *Coordinator) handleRequests() {
	msgChan := c.cluster.ReceiveChan()
	for {
		select {
		case msg := <-msgChan:
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
			return
		}
	}
}

func (c *Coordinator) handlePingRequest(msg *ClusterMsg) {
	log.Debugf("Received Ping request from %s", msg.Originator)
	msg.MsgType = PingResponse

	ctx, cancel := context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
	defer cancel()
	err := c.cluster.Reply(msg, ctx)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to send Ping response")
	}
}

func (c *Coordinator) handleVerifyNodeListRequest(msg *ClusterMsg) {
	log.Debugf("Received VerifyNodeList request from %s", msg.Originator)

}

// func (c *Coordinator) handleGetNodeListRequest(msg *ClusterMsg) {
// 	log.Debugf("Received GetNodeList request from %s", msg.Originator)
//
// 	replyChans := make(chan chan *ClusterMsg)
//
// 	// Send ping request to all nodes
// 	for _, node := range c.cluster.clusterConfig.Nodes {
// 		msg.MsgType = PingRequest
// 		msg.Recipient = node.Name
// 		ctx, cancel := context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
// 		defer cancel()
// 		replyChan, err := c.cluster.SendAndReceive(node.Name, msg, ctx)
// 		if err != nil {
// 			log.WithFields(log.Fields{"error": err}).Error("Failed to send Ping request")
// 		}
// 		replyChans <- replyChan
// 	}
//
// 	// Wait for responses
// 	ctx, cancel := context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
// 	defer cancel()
//
// 	pingResponses := make(chan *ClusterMsg)
// Loop:
// 	for {
// 		done := make(chan bool)
// 		select {
// 		case <-done:
// 			log.Debug("Received all ping responses")
// 			break Loop
// 		case <-ctx.Done():
// 			log.Error("Timeout waiting for ping responses")
// 			break Loop
// 		case replyChan := <-replyChans:
// 			go func() {
// 				ctx, cancel := context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
// 				defer cancel()
//
// 				select {
// 				case <-ctx.Done():
// 					log.Error("Timeout waiting for ping response")
// 					return
// 				case replyMsg := <-replyChan:
// 					if replyMsg.MsgType == PingResponse {
// 						log.Debugf("Received ping response from %s", replyMsg.Originator)
// 					} else {
// 						log.WithFields(log.Fields{"msgType": replyMsg.MsgType}).Error("Unexpected message type, expected ping response")
// 					}
// 					pingResponses <- replyMsg
// 					if len(pingResponses) == len(c.cluster.clusterConfig.Nodes) {
// 						done <- true
// 						return
// 					}
// 				}
// 			}()
// 		}
// 	}
//
// 	close(pingResponses)
//
// 	nodeList := make([]string, 0)
// 	for msg := range pingResponses {
// 		nodeList = append(nodeList, msg.Originator)
// 	}
//
// 	log.WithFields(log.Fields{"NodeList": nodeList}).Debug("Sending node list response")
// 	msg.MsgType = NodeListResponse
// 	msg.NodeList = nodeList
// 	ctx, cancel = context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
// 	defer cancel()
// 	err := c.cluster.Reply(msg, ctx)
// 	if err != nil {
// 		log.WithFields(log.Fields{"error": err}).Error("Failed to send NodeList response")
// 	}
// }

func (c *Coordinator) handleGetNodeListRequest(msg *ClusterMsg) {
	log.Debugf("Received GetNodeList request from %s", msg.Originator)

	replyChans := make(chan chan *ClusterMsg, len(c.cluster.clusterConfig.Nodes))

	// Send ping request to all nodes
	for _, node := range c.cluster.clusterConfig.Nodes {
		msgCopy := *msg // Copy the message to avoid modifying it during iteration
		msgCopy.MsgType = PingRequest
		msgCopy.Recipient = node.Name
		ctx, cancel := context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
		defer cancel()
		replyChan, err := c.cluster.SendAndReceive(node.Name, &msgCopy, ctx)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Failed to send Ping request")
		}
		replyChans <- replyChan
	}
	close(replyChans) // Close the replyChans channel when done sending

	// Wait for responses
	ctx, cancel := context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
	defer cancel()

	pingResponses := make(chan *ClusterMsg, len(c.cluster.clusterConfig.Nodes))
	var wg sync.WaitGroup // Use a WaitGroup to wait for all goroutines to finish

	for replyChan := range replyChans {
		wg.Add(1)
		go func(replyChan chan *ClusterMsg) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
			defer cancel()

			select {
			case <-ctx.Done():
				log.Error("Timeout waiting for ping response")
			case replyMsg := <-replyChan:
				if replyMsg.MsgType == PingResponse {
					log.Debugf("Received ping response from %s", replyMsg.Originator)
				} else {
					log.WithFields(log.Fields{"msgType": replyMsg.MsgType}).Error("Unexpected message type, expected ping response")
				}
				pingResponses <- replyMsg
			}
		}(replyChan)
	}

	go func() {
		wg.Wait()
		close(pingResponses) // Close pingResponses when all goroutines are done
	}()

	nodeList := make([]string, 0)
	for msg := range pingResponses {
		nodeList = append(nodeList, msg.Originator)
	}

	log.WithFields(log.Fields{"NodeList": nodeList}).Debug("Sending node list response")
	msg.MsgType = NodeListResponse
	msg.NodeList = nodeList
	ctx, cancel = context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
	defer cancel()
	err := c.cluster.Reply(msg, ctx)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to send NodeList response")
	}
}

func (c *Coordinator) handleRPCRequest(msg *ClusterMsg) {
	log.Debugf("Received RPC request from %s", msg.Originator)
}
