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

type Cluster struct {
	ginHandler             *gin.Engine
	restyClient            *resty.Client
	clusterConfig          Config
	thisNode               Node
	incomingClusterMsgChan chan *ClusterMsg
	replyChans             map[string]response
	receiveChan            chan []byte
	mutex                  *sync.Mutex
}

type response struct {
	reciveChan chan *ClusterMsg
	added      int64
}

func CreateCluster(thisNode Node, clusterConfig Config, ginHandler *gin.Engine) *Cluster {
	c := &Cluster{}
	c.ginHandler = ginHandler
	c.restyClient = resty.New()
	c.clusterConfig = clusterConfig
	c.thisNode = thisNode
	c.incomingClusterMsgChan = make(chan *ClusterMsg, 1000)
	c.mutex = &sync.Mutex{}
	c.replyChans = make(map[string]response)

	c.setupRoutes()

	return c
}

func (c *Cluster) setupRoutes() {
	c.ginHandler.POST("/cluster", c.handleClusterRequest)
}

func (c *Cluster) handleClusterRequest(ctx *gin.Context) {
	jsonBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		errMsg := "Bad relay request"
		log.WithFields(log.Fields{"Error": err}).Error(errMsg)
		ctx.String(http.StatusBadRequest, errMsg)
	}

	clusterMsg, err := DeserializeClusterMsg(jsonBytes)
	if err != nil {
		errMsg := "Malfomed cluster message"
		log.WithFields(log.Fields{"Error": err}).Error(errMsg)
		ctx.String(http.StatusBadRequest, errMsg)
	}

	msgID := clusterMsg.ID

	// Check if msgID is in replyChans
	c.mutex.Lock()
	defer c.mutex.Unlock()

	response, ok := c.replyChans[msgID]
	if ok {
		response.reciveChan <- clusterMsg
		delete(c.replyChans, msgID)
	} else {
		c.incomingClusterMsgChan <- clusterMsg
	}

	ctx.String(http.StatusOK, "")
}

func (c *Cluster) Send(name string, msg *ClusterMsg, ctx context.Context) error {
	return c.send(name, msg, false, ctx)
}

func (c *Cluster) Reply(msg *ClusterMsg, ctx context.Context) error {
	return c.send(msg.Originator, msg, true, ctx)
}

func (c *Cluster) send(name string, msg *ClusterMsg, reply bool, ctx context.Context) error {
	if msg == nil {
		log.WithFields(log.Fields{"Error": "Trying to send nil cluster message"}).Error("Nil cluster message")
		return errors.New("ClusterMsg is nil")
	}

	msg.Originator = c.thisNode.Name

	if !reply {
		msg.ID = core.GenerateRandomID()
	}

	buf, err := msg.Serialize()
	if err != nil {
		return err
	}

	found := false
	for _, node := range c.clusterConfig.Nodes {
		if node.Name == name {
			found = true
			_, err := c.restyClient.R().
				SetContext(ctx).
				SetBody(buf).
				Post("http://" + node.Host + ":" + strconv.Itoa(node.RelayPort) + "/cluster")
			if err != nil {
				return err
			}
		}
	}

	if !found {
		return errors.New("Node not found")
	}

	return nil
}

func (c *Cluster) SendAndReceive(name string, msg *ClusterMsg, ctx context.Context) (chan *ClusterMsg, error) {
	if msg == nil {
		log.WithFields(log.Fields{"Error": "Trying to send nil cluster message"}).Error("Nil cluster message")
		return nil, errors.New("ClusterMsg is nil")
	}

	msgID := core.GenerateRandomID()
	msg.ID = msgID
	msg.Originator = c.thisNode.Name

	buf, err := msg.Serialize()
	if err != nil {
		return nil, err
	}

	receiveChan := make(chan *ClusterMsg) // This need to be closed by the caller
	now := time.Now().Unix()
	response := response{reciveChan: receiveChan, added: now}
	c.mutex.Lock()
	c.replyChans[msgID] = response
	c.mutex.Unlock()

	found := false
	for _, node := range c.clusterConfig.Nodes {
		if node.Name == name {
			found = true
			_, err := c.restyClient.R().
				SetContext(ctx).
				SetBody(buf).
				Post("http://" + node.Host + ":" + strconv.Itoa(node.RelayPort) + "/cluster")
			if err != nil {
				return nil, err
			}
		}
	}

	if !found {
		return nil, errors.New("Node not found")
	}

	return receiveChan, nil
}

func (c *Cluster) ReceiveChan() chan *ClusterMsg {
	return c.incomingClusterMsgChan
}
