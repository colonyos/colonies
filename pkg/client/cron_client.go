package client

import (
	"context"

	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/rpc"
)

func (client *ColoniesClient) AddCron(cron *core.Cron, prvKey string) (*core.Cron, error) {
	msg := rpc.CreateAddCronMsg(cron)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.AddCronPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToCron(respBodyString)
}

func (client *ColoniesClient) GetCron(cronID string, prvKey string) (*core.Cron, error) {
	msg := rpc.CreateGetCronMsg(cronID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetCronPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToCron(respBodyString)
}

func (client *ColoniesClient) GetCrons(colonyName string, count int, prvKey string) ([]*core.Cron, error) {
	msg := rpc.CreateGetCronsMsg(colonyName, count)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.GetCronsPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToCronArray(respBodyString)
}

func (client *ColoniesClient) RunCron(cronID string, prvKey string) (*core.Cron, error) {
	msg := rpc.CreateRunCronMsg(cronID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return nil, err
	}

	respBodyString, err := client.sendMessage(rpc.RunCronPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return nil, err
	}

	return core.ConvertJSONToCron(respBodyString)
}

func (client *ColoniesClient) RemoveCron(cronID string, prvKey string) error {
	msg := rpc.CreateRemoveCronMsg(cronID)
	jsonString, err := msg.ToJSON()
	if err != nil {
		return err
	}

	_, err = client.sendMessage(rpc.RemoveCronPayloadType, jsonString, prvKey, false, context.TODO())
	if err != nil {
		return err
	}

	return nil
}