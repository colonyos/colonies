package client

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
	log "github.com/sirupsen/logrus"
)

func (client *ColoniesClient) SubscribeProcesses(colonyName string, executorType string, state int, timeout int, prvKey string) (*ProcessSubscription, error) {
	log.WithFields(log.Fields{"ColonyName": colonyName, "ExecutorType": executorType, "State": state, "Timeout": timeout}).Debug("SubscribeProcesses called")

	msg := rpc.CreateSubscribeProcessesMsg(colonyName, executorType, state, timeout)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	rpcMsg, err := rpc.CreateRPCMsg(rpc.SubscribeProcessesPayloadType, jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	jsonString, err = rpcMsg.ToJSON()
	if err != nil {
		return nil, err
	}

	conn, err := client.establishRealtimeConn(jsonString)
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Debug("SubscribeProcesses: failed to establish realtime connection")
		return nil, err
	}

	log.Debug("SubscribeProcesses: WebSocket connection established, waiting for messages")

	subscription := createProcessSubscription(conn)
	go func(subscription *ProcessSubscription) {
		for {
			_, jsonBytes, err := subscription.conn.ReadMessage()
			if err != nil {
				log.WithFields(log.Fields{"Error": err}).Debug("SubscribeProcesses: read error, closing")
				subscription.ErrChan <- err
				return
			}

			log.WithFields(log.Fields{"Size": len(jsonBytes)}).Debug("SubscribeProcesses: received message")

			rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(string(jsonBytes))
			if err != nil {
				subscription.ErrChan <- err
				continue
			}

			if rpcReplyMsg.Error {
				failureMsg, err := core.ConvertJSONToFailure(rpcReplyMsg.DecodePayload())
				if err != nil {
					subscription.ErrChan <- err
				}
				subscription.ErrChan <- errors.New(failureMsg.Message)
			}

			process, err := core.ConvertJSONToProcess(rpcReplyMsg.DecodePayload())
			if err != nil {
				subscription.ErrChan <- err
				continue
			}

			log.WithFields(log.Fields{"ProcessID": process.ID, "State": process.State}).Debug("SubscribeProcesses: received process")
			subscription.ProcessChan <- process
		}
	}(subscription)

	return subscription, nil
}

func (client *ColoniesClient) SubscribeProcess(colonyName string, processID string, executorType string, state int, timeout int, prvKey string) (*ProcessSubscription, error) {
	log.WithFields(log.Fields{"ColonyName": colonyName, "ProcessID": processID, "ExecutorType": executorType, "State": state, "Timeout": timeout}).Debug("SubscribeProcess called")

	msg := rpc.CreateSubscribeProcessMsg(colonyName, processID, executorType, state, timeout)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	rpcMsg, err := rpc.CreateRPCMsg(rpc.SubscribeProcessPayloadType, jsonString, prvKey)
	if err != nil {
		return nil, err
	}

	jsonString, err = rpcMsg.ToJSON()
	if err != nil {
		return nil, err
	}

	conn, err := client.establishRealtimeConn(jsonString)
	if err != nil {
		log.WithFields(log.Fields{"Error": err, "ProcessID": processID}).Debug("SubscribeProcess: failed to establish realtime connection")
		return nil, err
	}

	log.WithFields(log.Fields{"ProcessID": processID}).Debug("SubscribeProcess: WebSocket connection established, waiting for messages")

	subscription := createProcessSubscription(conn)
	go func(subscription *ProcessSubscription) {
		for {
			_, jsonBytes, err := subscription.conn.ReadMessage()
			if err != nil {
				log.WithFields(log.Fields{"Error": err, "ProcessID": processID}).Debug("SubscribeProcess: read error, closing")
				subscription.ErrChan <- err
				return
			}

			log.WithFields(log.Fields{"ProcessID": processID, "Size": len(jsonBytes)}).Debug("SubscribeProcess: received message")

			rpcReplyMsg, err := rpc.CreateRPCReplyMsgFromJSON(string(jsonBytes))
			if err != nil {
				subscription.ErrChan <- err
				continue
			}

			if rpcReplyMsg.Error {
				failureMsg, err := core.ConvertJSONToFailure(rpcReplyMsg.DecodePayload())
				if err != nil {
					subscription.ErrChan <- err
				}
				subscription.ErrChan <- errors.New(failureMsg.Message)
			}

			process, err := core.ConvertJSONToProcess(rpcReplyMsg.DecodePayload())
			if err != nil {
				subscription.ErrChan <- err
				continue
			}

			log.WithFields(log.Fields{"ProcessID": process.ID, "State": process.State}).Debug("SubscribeProcess: received process update")
			subscription.ProcessChan <- process
		}
	}(subscription)

	return subscription, nil
}