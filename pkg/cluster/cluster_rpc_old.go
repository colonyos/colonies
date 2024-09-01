package cluster

//
// import (
// 	"context"
// 	"errors"
// 	"io"
// 	"net/http"
// 	"strconv"
// 	"sync"
// 	"time"
//
// 	"github.com/colonyos/colonies/pkg/core"
// 	"github.com/gin-gonic/gin"
// 	"github.com/go-resty/resty/v2"
// 	log "github.com/sirupsen/logrus"
// )
//
// type clusterRPC struct {
// 	ginHandler             *gin.Engine
// 	restyClient            *resty.Client
// 	clusterConfig          Config
// 	thisNode               Node
// 	incomingClusterMsgChan chan *ClusterMsg
// 	replyChans             map[string]*response
// 	mutex                  *sync.Mutex
// }
//
// type response struct {
// 	receiveChan chan *ClusterMsg
// 	msgID       string
// 	added       int64 // TODO: purge old messages after a certain time
// }
//
// func createClusterRPC(thisNode Node, clusterConfig Config, ginHandler *gin.Engine) *clusterRPC {
// 	rpc := &clusterRPC{}
// 	rpc.ginHandler = ginHandler
// 	rpc.restyClient = resty.New()
// 	rpc.clusterConfig = clusterConfig
// 	rpc.thisNode = thisNode
// 	rpc.incomingClusterMsgChan = make(chan *ClusterMsg, 1000)
// 	rpc.mutex = &sync.Mutex{}
// 	rpc.replyChans = make(map[string]*response)
//
// 	rpc.setupRoutes()
//
// 	return rpc
// }
//
// func (rpc *clusterRPC) setupRoutes() {
// 	rpc.ginHandler.POST("/cluster", rpc.handleClusterRequest)
// 	log.Debug("ClusterRPC: routes setup completed")
// }
//
// func (rpc *clusterRPC) handleClusterRequest(ctx *gin.Context) {
// 	jsonBytes, err := io.ReadAll(ctx.Request.Body)
// 	if err != nil {
// 		errMsg := "Bad relay request"
// 		log.WithFields(log.Fields{"Error": err}).Error(errMsg)
// 		ctx.String(http.StatusBadRequest, errMsg)
// 	}
//
// 	clusterMsg, err := DeserializeClusterMsg(jsonBytes)
// 	if err != nil {
// 		errMsg := "Malfomed cluster message"
// 		log.WithFields(log.Fields{"Error": err}).Error(errMsg)
// 		ctx.String(http.StatusBadRequest, errMsg)
// 	}
//
// 	msgID := clusterMsg.ID
//
// 	log.WithFields(log.Fields{"Node": rpc.thisNode.Name, "msgID": msgID}).Debug("ClusterRPC: Received cluster rpc message")
//
// 	// Check if msgID is in replyChans
// 	rpc.mutex.Lock()
// 	defer rpc.mutex.Unlock()
//
// 	response, ok := rpc.replyChans[msgID]
// 	if ok {
// 		response.receiveChan <- clusterMsg
// 		delete(rpc.replyChans, msgID)
// 	} else {
// 		rpc.incomingClusterMsgChan <- clusterMsg
// 	}
//
// 	log.WithFields(log.Fields{"msgID": msgID}).Debug("ClusterRPC: Adding msg to receive chan")
//
// 	ctx.String(http.StatusOK, "")
// }
//
// func (rpc *clusterRPC) send(name string, msg *ClusterMsg, ctx context.Context) error {
// 	return rpc.sendInternal(name, msg, false, ctx)
// }
//
// func (rpc *clusterRPC) reply(msg *ClusterMsg, ctx context.Context) error {
// 	return rpc.sendInternal(msg.Originator, msg, true, ctx)
// }
//
// func (rpc *clusterRPC) sendInternal(name string, msg *ClusterMsg, reply bool, ctx context.Context) error {
// 	if msg == nil {
// 		log.WithFields(log.Fields{"Error": "Trying to send nil cluster message"}).Error("ClusterRPC: Nil cluster message")
// 		return errors.New("ClusterMsg is nil")
// 	}
//
// 	msg.Originator = rpc.thisNode.Name
//
// 	if !reply {
// 		msg.ID = core.GenerateRandomID()
// 	}
//
// 	buf, err := msg.Serialize()
// 	if err != nil {
// 		return err
// 	}
//
// 	found := false
// 	for _, node := range rpc.clusterConfig.Nodes {
// 		if node.Name == name {
// 			found = true
// 			_, err := rpc.restyClient.R().
// 				SetContext(ctx).
// 				SetBody(buf).
// 				Post("http://" + node.Host + ":" + strconv.Itoa(node.RelayPort) + "/cluster")
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
//
// 	if !found {
// 		return errors.New("Node not found")
// 	}
//
// 	return nil
// }
//
// func (rpc *clusterRPC) sendAndReceive(name string, msg *ClusterMsg, ctx context.Context) (*response, error) {
// 	if msg == nil {
// 		log.WithFields(log.Fields{"Error": "Trying to send nil cluster message"}).Error("ClusterRPC: Nil cluster message")
// 		return nil, errors.New("ClusterMsg is nil")
// 	}
//
// 	msgID := core.GenerateRandomID()
// 	msg.ID = msgID
// 	msg.Originator = rpc.thisNode.Name
//
// 	buf, err := msg.Serialize()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	receiveChan := make(chan *ClusterMsg) // NOTE! This need to be closed by the caller
// 	now := time.Now().Unix()
// 	response := &response{receiveChan: receiveChan, msgID: msgID, added: now}
// 	rpc.mutex.Lock()
// 	rpc.replyChans[msgID] = response
// 	rpc.mutex.Unlock()
//
// 	found := false
// 	for _, node := range rpc.clusterConfig.Nodes {
// 		if node.Name == name {
// 			found = true
// 			_, err := rpc.restyClient.R().
// 				SetContext(ctx).
// 				SetBody(buf).
// 				Post("http://" + node.Host + ":" + strconv.Itoa(node.RelayPort) + "/cluster")
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 	}
//
// 	if !found {
// 		return nil, errors.New("Node not found")
// 	}
//
// 	return response, nil
// }
//
// func (rpc *clusterRPC) receiveChan() chan *ClusterMsg {
// 	return rpc.incomingClusterMsgChan
// }
//
// func (rpc *clusterRPC) close(r *response) {
// 	if r == nil {
// 		log.WithFields(log.Fields{"Error": "Trying to close nil response"}).Error("ClusterRPC: Nil response")
// 		return
// 	}
// 	rpc.mutex.Lock()
// 	defer rpc.mutex.Unlock()
// 	close(r.receiveChan)
// 	delete(rpc.replyChans, r.msgID)
//
// 	log.WithFields(log.Fields{"PendingResponses": len(rpc.replyChans)}).Debug("ClusterRPC: Cleaning up")
// }
