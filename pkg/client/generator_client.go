package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) AddGenerator(generator *core.Generator, prvKey string) (*core.Generator, error) {
	msg := rpc.CreateAddGeneratorMsg(generator)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddGeneratorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToGenerator(respBodyString)
}

func (client *ColoniesClient) GetGenerator(generatorID string, prvKey string) (*core.Generator, error) {
	msg := rpc.CreateGetGeneratorMsg(generatorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetGeneratorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToGenerator(respBodyString)
}

func (client *ColoniesClient) ResolveGenerator(colonyName string, generatorName string, prvKey string) (*core.Generator, error) {
	msg := rpc.CreateResolveGeneratorMsg(colonyName, generatorName)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.ResolveGeneratorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToGenerator(respBodyString)
}

func (client *ColoniesClient) GetGenerators(colonyName string, count int, prvKey string) ([]*core.Generator, error) {
	msg := rpc.CreateGetGeneratorsMsg(colonyName, count)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetGeneratorsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToGeneratorArray(respBodyString)
}

func (client *ColoniesClient) PackGenerator(generatorID string, arg string, prvKey string) error {
	msg := rpc.CreatePackGeneratorMsg(generatorID, arg)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.PackGeneratorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}

func (client *ColoniesClient) RemoveGenerator(generatorID string, prvKey string) error {
	msg := rpc.CreateRemoveGeneratorMsg(generatorID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveGeneratorPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}