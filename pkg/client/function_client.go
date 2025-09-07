package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) AddFunction(function *core.Function, prvKey string) (*core.Function, error) {
	msg := rpc.CreateAddFunctionMsg(function)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddFunctionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFunction(respBodyString)
}

func (client *ColoniesClient) GetFunctionsByExecutor(colonyName string, executorName string, prvKey string) ([]*core.Function, error) {
	msg := rpc.CreateGetFunctionsMsg(colonyName, executorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFunctionsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFunctionArray(respBodyString)
}

func (client *ColoniesClient) GetFunctionsByColony(colonyName string, prvKey string) ([]*core.Function, error) {
	msg := rpc.CreateGetFunctionsByColonyNameMsg(colonyName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetFunctionsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToFunctionArray(respBodyString)
}

func (client *ColoniesClient) RemoveFunction(functionID string, prvKey string) error {
	msg := rpc.CreateRemoveFunctionMsg(functionID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveFunctionPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}