package cluster

//
// import (
// 	"context"
//
// 	"github.com/colonyos/colonies/pkg/core"
// 	log "github.com/sirupsen/logrus"
// )
//
// const MAX_RESPONSE_TIME = 1 // Wait max 1 second for a response from a replica
//
// type Coordinator struct {
// 	coordinator *Coordinator
// 	etcdServer  *EtcdServer
// 	cluster     *Cluster
// 	doneChan    chan bool
// }
//
// func CreateCoordinator(etcdServer *EtcdServer, cluster *Cluster) *Coordinator {
// 	return &Coordinator{etcdServer: etcdServer, cluster: cluster, doneChan: make(chan bool)}
// }
//
// func (cs *Coordinator) sendMsg(msg *ClusterMsg, ctx context.Context) error {
// 	err := cs.cluster.Send(msg.Recipient, msg, ctx)
// 	if err != nil {
// 		log.WithFields(log.Fields{"Error": err}).Error("Failed to send cluster message")
// 		return err
// 	}
//
// 	log.WithFields(log.Fields{"Message": msg}).Debug("Sent cluster message")
// 	return nil
// }
//
// func (cs *Coordinator) deserializeIncomingMsg(rawMsg []byte) (*ClusterMsg, error) {
// 	msg, err := DeserializeClusterMsg(rawMsg)
// 	if err != nil {
// 		log.WithFields(log.Fields{"Error": err}).Error("Failed to deserialize cluster message")
// 		return nil, err
// 	}
//
// 	return msg, nil
// }
//
// func (cs *Coordinator) handlePingRequest(msg *ClusterMsg) {
// 	replyMsg := &ClusterMsg{
// 		MsgType:    PingResponseMsgType,
// 		ID:         msg.ID,
// 		Originator: cs.cluster.thisNode.Name,
// 		Recipient:  msg.Originator}
//
// 	err := cs.sendMsg(replyMsg, context.Background())
// 	if err != nil {
// 		log.WithFields(log.Fields{"Error": err}).Error("Failed to send ping response")
// 	}
// }
//
// func (cs *Coordinator) RunForever() {
// 	receiveChan := cs.cluster.ReceiveChan()
// 	for {
// 		select {
// 		case <-cs.doneChan:
// 			log.Info("Shutting down coordinator")
// 			return
// 		case rawMsg := <-receiveChan:
// 			msg, err := cs.deserializeIncomingMsg(rawMsg)
// 			if err != nil {
// 				continue
// 			}
//
// 			switch msg.MsgType {
// 			case PingRequestMsgType:
// 				cs.handlePingRequest(msg)
// 				break
// 			}
// 		}
// 	}
// }
//
// func (cs *Coordinator) pingAllNodes() {
// 	ctx, _ := context.WithTimeout(context.Background(), MAX_RESPONSE_TIME)
//
// 	msgID := core.GenerateRandomID()
//
// 	receiveChan := make(chan *ClusterMsg)
// 	//now := time.Now().Unix()
// 	// response := response{reciveChan: receiveChan, added: now}
// 	// cs.mutex.Lock()
// 	// cs.reciveChans[msgID] = response
// 	// cs.mutex.Unlock()
//
// 	nodes := cs.cluster.clusterConfig.Nodes
// 	for _, node := range nodes {
// 		msg := &ClusterMsg{
// 			MsgType:    PingRequestMsgType,
// 			ID:         msgID,
// 			Originator: cs.cluster.thisNode.Name,
// 			Recipient:  node.Name}
//
// 		go func() {
// 			err := cs.cluster.Send(node.Name, msg, ctx)
// 			if err != nil {
// 				log.WithFields(log.Fields{"Error": err, "Node": node}).Error("Failed to send ping request")
// 			}
// 		}()
// 	}
//
// 	for {
// 		select {
// 		case msg := <-receiveChan:
// 			log.WithFields(log.Fields{"From": msg.Originator}).Debug("Received ping response")
// 			break
// 		case <-ctx.Done():
// 			log.Error("Timeout waiting for ping responses")
// 			break
// 		}
// 	}
// }
//
// func Send(resourceID string, msg *ClusterMsg) error {
// 	// TODO
//
// 	// 1. Get the cluster leader ID
// 	// 2. Send a message to the leader to check that we have the latest cluster configuration
// 	//    If not, ask the leader to send the latest configuration
// 	// 3. Use consistent hashing to find the target server
// 	// 4. Send the message to the target server
//
// 	return nil
// }
//
// func (cs *Coordinator) Shutdown() {
// 	cs.doneChan <- true
// }
