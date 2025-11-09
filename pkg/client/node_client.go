package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) GetNodes(colonyName string, prvKey string) ([]*core.Node, error) {
	msg := rpc.CreateGetNodesMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetNodesPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToNodes(respBodyString)
}

func (client *ColoniesClient) GetNode(colonyName string, nodeName string, prvKey string) (*core.Node, error) {
	msg := rpc.CreateGetNodeMsg(colonyName, nodeName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetNodePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToNode(respBodyString)
}

func (client *ColoniesClient) GetNodesByLocation(colonyName string, location string, prvKey string) ([]*core.Node, error) {
	msg := rpc.CreateGetNodesByLocationMsg(colonyName, location)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetNodesByLocationPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToNodes(respBodyString)
}
