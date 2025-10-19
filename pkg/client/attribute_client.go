package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) AddAttribute(attribute core.Attribute, prvKey string) (core.Attribute, error) {
	msg := rpc.CreateAddAttributeMsg(attribute)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return core.Attribute{}, err
	}

	respBodyString, err := client.sendMessage(rpc.AddAttributePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return core.Attribute{}, err
	}

	return core.ConvertJSONToAttribute(respBodyString)
}

func (client *ColoniesClient) GetAttribute(attributeID string, prvKey string) (core.Attribute, error) {
	msg := rpc.CreateGetAttributeMsg(attributeID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return core.Attribute{}, err
	}

	respBodyString, err := client.sendMessage(rpc.GetAttributePayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return core.Attribute{}, err
	}

	return core.ConvertJSONToAttribute(respBodyString)
}