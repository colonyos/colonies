package cluster

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

type clusterRPC struct {
	ginHandler             *gin.Engine
	restyClient            *resty.Client
	clusterConfig          *Config
	thisNode               *Node
	incomingClusterMsgChan chan *ClusterMsg
	pendingResponses       map[string]*response
	mutex                  *sync.Mutex
	doneChan               chan struct{}
}

type response struct {
	receiveChan chan *ClusterMsg
	errChan     chan error
	msgID       string
	added       int64 // Used to purge old messages after a certain time
}

func createClusterRPC(thisNode *Node, clusterConfig *Config, ginHandler *gin.Engine, purgeInterval time.Duration) *clusterRPC {
	rpc := &clusterRPC{
		ginHandler:             ginHandler,
		restyClient:            resty.New(),
		clusterConfig:          clusterConfig,
		thisNode:               thisNode,
		incomingClusterMsgChan: make(chan *ClusterMsg, 1000),
		mutex:                  &sync.Mutex{},
		pendingResponses:       make(map[string]*response),
		doneChan:               make(chan struct{}),
	}

	rpc.setupRoutes()

	// Periodically clean up old responses, in cases a node crashes before cleaning up
	go rpc.cleanupOldResponses(purgeInterval)

	return rpc
}

func (rpc *clusterRPC) setupRoutes() {
	rpc.ginHandler.POST("/cluster", rpc.handleClusterRequest)
	log.Debug("ClusterRPC: routes setup completed")
}

func (rpc *clusterRPC) handleClusterRequest(ctx *gin.Context) {
	jsonBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		errMsg := "Bad relay request"
		log.WithFields(log.Fields{"Error": err}).Error(errMsg)
		ctx.String(http.StatusBadRequest, errMsg)
		return // Early return on error
	}

	clusterMsg, err := DeserializeClusterMsg(jsonBytes)
	if err != nil {
		errMsg := "Malformed cluster message"
		log.WithFields(log.Fields{"Error": err}).Error(errMsg)
		ctx.String(http.StatusBadRequest, errMsg)
		return // Early return on error
	}

	msgID := clusterMsg.ID

	log.WithFields(log.Fields{"Node": rpc.thisNode.Name, "MsgID": msgID, "MagType": clusterMsg.MsgType}).Debug("ClusterRPC: Received cluster rpc message")

	rpc.mutex.Lock()
	response, ok := rpc.pendingResponses[msgID]
	if ok {
		response.receiveChan <- clusterMsg
		delete(rpc.pendingResponses, msgID)
	} else {
		rpc.incomingClusterMsgChan <- clusterMsg
	}
	rpc.mutex.Unlock()

	log.WithFields(log.Fields{"msgID": msgID}).Debug("ClusterRPC: Added msg to receive chan")

	ctx.String(http.StatusOK, "")
}

func (rpc *clusterRPC) send(name string, msg *ClusterMsg) {
	rpc.sendInternal(name, msg, false)
}

func (rpc *clusterRPC) reply(msg *ClusterMsg) {
	rpc.sendInternal(msg.Originator, msg, true)
}

func (rpc *clusterRPC) sendInternal(name string, msg *ClusterMsg, reply bool) {
	if msg == nil {
		log.WithFields(log.Fields{"Error": "Trying to send nil cluster message"}).Error("ClusterRPC: Nil cluster message")
		return
	}

	log.WithFields(log.Fields{"Receiver": name, "MsgType": msg.MsgType}).Debug("ClusterRPC: Sending cluster rpc message")

	go func(msgCopy ClusterMsg) {
		ctx, cancel := context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
		defer cancel()

		msgCopy.Originator = rpc.thisNode.Name

		if !reply {
			msgCopy.ID = core.GenerateRandomID()
		}

		buf, err := msgCopy.Serialize()
		if err != nil {
			log.WithFields(log.Fields{"Error": err}).Error("ClusterRPC: Error serializing message")
			return
		}

		found := false
		for _, node := range rpc.clusterConfig.Nodes {
			if node.Name == name {
				found = true
				_, err := rpc.restyClient.R().
					SetContext(ctx).
					SetBody(buf).
					Post("http://" + node.Host + ":" + strconv.Itoa(node.RelayPort) + "/cluster")
				if err != nil {
					log.WithFields(log.Fields{"Error": err}).Error("ClusterRPC: Error sending message 2")
					return
				}
			}
		}

		if !found {
			log.WithFields(log.Fields{"Error": "Node not found"}).Error("ClusterRPC: Node not found")
			return
		}
	}(*msg)
}

func (rpc *clusterRPC) sendAndReceive(name string, msg *ClusterMsg) (*response, error) {
	if msg == nil {
		log.WithFields(log.Fields{"Error": "Trying to send nil cluster message"}).Error("ClusterRPC: Nil cluster message")
		return nil, errors.New("ClusterMsg is nil")
	}

	if rpc.clusterConfig.Nodes == nil {
		return nil, errors.New("Cluster not configured")
	}

	if len(rpc.clusterConfig.Nodes) == 0 {
		return nil, errors.New("Cluster not configured")
	}

	if rpc.thisNode.Name == name {
		return nil, errors.New("Cannot send message to self")
	}

	if rpc.thisNode.Name == msg.Originator {
		return nil, errors.New("Cannot send message to self")
	}

	node := rpc.clusterConfig.Nodes[name]
	if node == nil {
		return nil, errors.New("Node not found")
	}

	msgID := core.GenerateRandomID()
	msg.ID = msgID
	msg.Originator = rpc.thisNode.Name

	buf, err := msg.Serialize()
	if err != nil {
		return nil, err
	}

	receiveChan := make(chan *ClusterMsg)
	errChan := make(chan error)
	now := time.Now().Unix()
	resp := &response{receiveChan: receiveChan, errChan: errChan, msgID: msgID, added: now}

	rpc.mutex.Lock()
	rpc.pendingResponses[msgID] = resp
	rpc.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), PING_RESPONSE_TIMEOUT*time.Second)
	go func(resp *response, host string, port int, buf []byte) {
		defer cancel()
		_, err := rpc.restyClient.R().
			SetContext(ctx).
			SetBody(buf).
			Post("http://" + host + ":" + strconv.Itoa(port) + "/cluster")
		if err != nil {
			resp.errChan <- err
			log.WithFields(log.Fields{"Error": err}).Error("ClusterRPC: Error sending message")
		}
	}(resp, node.Host, node.RelayPort, buf)

	return resp, nil
}

func (rpc *clusterRPC) receiveChan() chan *ClusterMsg {
	return rpc.incomingClusterMsgChan
}

func (rpc *clusterRPC) close(r *response) {
	if r == nil {
		log.WithFields(log.Fields{"Error": "Trying to close nil response"}).Error("ClusterRPC: Nil response")
		return
	}

	rpc.mutex.Lock()
	defer rpc.mutex.Unlock()

	if _, exists := rpc.pendingResponses[r.msgID]; exists {
		close(r.receiveChan)
		close(r.errChan)
		delete(rpc.pendingResponses, r.msgID)
		log.WithFields(log.Fields{"PendingResponses": len(rpc.pendingResponses)}).Debug("ClusterRPC: Cleaning up")
	}
}

func (rpc *clusterRPC) shutdown() {
	rpc.doneChan <- struct{}{}
	close(rpc.doneChan)
}

func (rpc *clusterRPC) cleanupOldResponses(purgeInterval time.Duration) {
	log.WithFields(log.Fields{"Interval": purgeInterval}).Debug("ClusterRPC: Starting cleanup of old responses")
	ticker := time.NewTicker(purgeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Debug("ClusterRPC: Cleaning up old pending responses now")
			cutoff := time.Now().Unix() - int64(purgeInterval.Seconds())

			rpc.mutex.Lock()
			for msgID, resp := range rpc.pendingResponses {
				log.WithFields(log.Fields{
					"msgID":       msgID,
					"AddedTime":   resp.added,
					"CutoffTime":  cutoff,
					"ShouldPurge": resp.added < cutoff,
				}).Debug("ClusterRPC: Checking if response should be purged")

				if resp.added < cutoff {
					close(resp.receiveChan)
					close(resp.errChan)
					delete(rpc.pendingResponses, resp.msgID)
					log.WithFields(log.Fields{"PendingResponses": len(rpc.pendingResponses)}).Debug("ClusterRPC: Cleaning up")
				}
			}
			rpc.mutex.Unlock()

		case <-rpc.doneChan:
			log.Debug("ClusterRPC: Stopping cleanup of old responses")
			return
		}
	}
}
