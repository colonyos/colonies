package client

import (
	"errors"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) SubscribeProcesses(colonyName string, executorType string, state int, timeout int, prvKey string) (*ProcessSubscription, error) {
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
		return nil, err
	}

	subscription := createProcessSubscription(conn)
	go func(subscription *ProcessSubscription) {
		for {
			_, jsonBytes, err := subscription.conn.ReadMessage()
			if err != nil {
				subscription.ErrChan <- err
				continue
			}

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

			subscription.ProcessChan <- process
		}
	}(subscription)

	return subscription, nil
}

func (client *ColoniesClient) SubscribeProcess(colonyName string, processID string, executorType string, state int, timeout int, prvKey string) (*ProcessSubscription, error) {
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
		return nil, err
	}

	subscription := createProcessSubscription(conn)
	go func(subscription *ProcessSubscription) {
		for {
			_, jsonBytes, err := subscription.conn.ReadMessage()
			if err != nil {
				subscription.ErrChan <- err
				continue
			}

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

			subscription.ProcessChan <- process
		}
	}(subscription)

	return subscription, nil
}